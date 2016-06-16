// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/models"
)

//AssertError asserts that the specified error is not nil
func AssertError(g *G, err error) {
	g.Assert(err == nil).IsFalse("Expected error to exist, but it was nil")
}

//AssertNotError asserts that the specified error is nil
func AssertNotError(g *G, err error) {
	if err != nil {
		g.Assert(err == nil).IsTrue(err.Error())
	}
}

func TestClanHandler(t *testing.T) {
	g := Goblin(t)
	testDb, err := models.GetTestDB()
	AssertNotError(g, err)

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Create Clan Handler", func() {
		g.It("Should create clan", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			AssertNotError(g, err)

			gameID := player.GameID
			publicID := randomdata.FullName(randomdata.RandomGender)
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PostJSON(a, GetGameRoute(gameID, "/clans"), t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, gameID, publicID)
			AssertNotError(g, err)
			g.Assert(dbClan.GameID).Equal(gameID)
			g.Assert(dbClan.PublicID).Equal(publicID)
			g.Assert(dbClan.Name).Equal(clanName)
			g.Assert(dbClan.OwnerID).Equal(player.ID)
			g.Assert(dbClan.Metadata).Equal(metadata)
		})

		g.It("Should not create clan if invalid payload", func() {
			a := GetDefaultTestApp()
			gameID := "gameID"
			res := PostBody(a, GetGameRoute(gameID, "/clans"), t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create clan if owner does not exist", func() {
			gameID := randomdata.FullName(randomdata.RandomGender)
			publicID := randomdata.FullName(randomdata.RandomGender)
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := randomdata.FullName(randomdata.RandomGender)
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PostJSON(a, GetGameRoute(gameID, "/clans"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Player was not found with id: %s", ownerPublicID))
		})

		g.It("Should not create clan if invalid data", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			AssertNotError(g, err)

			gameID := player.GameID
			publicID := randomdata.FullName(randomdata.RandomGender)
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "it-will-fail-beacause-metada-is-not-a-json"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"PublicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			res := PostJSON(a, GetGameRoute(gameID, "/clans"), t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})

	g.Describe("Leave Clan Handler", func() {
		g.It("Should leave a clan and transfer ownership", func() {
			clan, owner, _, memberships, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()
			ownerPublicID := owner.PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"ownerPublicID": ownerPublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/leave", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			AssertNotError(g, err)
			g.Assert(dbClan.OwnerID).Equal(memberships[0].PlayerID)
		})

		g.It("Should not leave a clan if invalid payload", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/leave", "random-id"))
			res := PostBody(a, route, t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not leave a clan if player is not the owner", func() {
			clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"ownerPublicID": players[0].PublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/leave", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})
	})

	g.Describe("Transfer Clan Ownership Handler", func() {
		g.It("Should transfer a clan ownership", func() {
			clan, owner, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()
			ownerPublicID := owner.PublicID
			playerPublicID := players[0].PublicID

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"ownerPublicID":  ownerPublicID,
				"playerPublicID": playerPublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/transfer-ownership", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, clan.GameID, clan.PublicID)
			AssertNotError(g, err)
			g.Assert(dbClan.OwnerID).Equal(players[0].ID)
		})

		g.It("Should not transfer a clan ownership if invalid payload", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("clans/%s/transfer-ownership", "random-id"))
			res := PostBody(a, route, t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not transfer a clan ownership if player is not the owner", func() {
			clan, _, players, _, err := models.GetClanWithMemberships(testDb, 1, "", "")
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"ownerPublicID":  players[0].PublicID,
				"playerPublicID": players[0].PublicID,
			}
			route := GetGameRoute(clan.GameID, fmt.Sprintf("clans/%s/transfer-ownership", clan.PublicID))
			res := PostJSON(a, route, t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})
	})

	g.Describe("Update Clan Handler", func() {
		g.It("Should update clan", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			AssertNotError(g, err)

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "{\"new\": \"metadata\"}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			res := PutJSON(a, route, t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbClan, err := models.GetClanByPublicID(a.Db, gameID, publicID)
			AssertNotError(g, err)
			g.Assert(dbClan.GameID).Equal(gameID)
			g.Assert(dbClan.PublicID).Equal(publicID)
			g.Assert(dbClan.Name).Equal(clanName)
			g.Assert(dbClan.OwnerID).Equal(player.ID)
			g.Assert(dbClan.Metadata).Equal(metadata)
		})

		g.It("Should not update clan if invalid payload", func() {
			a := GetDefaultTestApp()

			route := GetGameRoute("game-id", fmt.Sprintf("/clans/%s", "random-id"))
			res := PutBody(a, route, t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not update clan if player is not the owner", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			AssertNotError(g, err)

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := randomdata.FullName(randomdata.RandomGender)
			metadata := "{\"x\": 1}"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			res := PutJSON(a, route, t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal(fmt.Sprintf("Clan was not found with id: %s", clan.PublicID))
		})

		g.It("Should not update clan if invalid data", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err := testDb.Insert(player)
			AssertNotError(g, err)

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":  player.GameID,
				"OwnerID": player.ID,
			}).(*models.Clan)
			err = testDb.Insert(clan)
			AssertNotError(g, err)

			gameID := clan.GameID
			publicID := clan.PublicID
			clanName := randomdata.FullName(randomdata.RandomGender)
			ownerPublicID := player.PublicID
			metadata := "it-will-fail-beacause-metada-is-not-a-json"

			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"gameID":        gameID,
				"publicID":      publicID,
				"name":          clanName,
				"ownerPublicID": ownerPublicID,
				"metadata":      metadata,
			}
			route := GetGameRoute(gameID, fmt.Sprintf("/clans/%s", publicID))
			res := PutJSON(a, route, t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})

	g.Describe("List All Clans Handler", func() {
		g.It("Should get all clans", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err = testDb.Insert(player)
			AssertNotError(g, err)

			expectedClans := []*models.Clan{}
			for i := 0; i < 10; i++ {
				clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
					"GameID":  player.GameID,
					"OwnerID": player.ID,
				}).(*models.Clan)
				err = testDb.Insert(clan)
				AssertNotError(g, err)
				expectedClans = append(expectedClans, clan)
			}
			sort.Sort(models.ClanByName(expectedClans))

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(player.GameID, "/clans"), t)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()
			for index, clanObj := range result["clans"].([]interface{}) {
				clan := clanObj.(map[string]interface{})
				g.Assert(clan["name"]).Equal(expectedClans[index].Name)
				g.Assert(clan["metadata"]).Equal(expectedClans[index].Metadata)
				g.Assert(clan["publicID"]).Equal(expectedClans[index].PublicID)
				g.Assert(clan["ID"]).Equal(nil)
			}
		})

		g.It("Should return empty list if invalid game query", func() {
			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute("invalid-query-game-id", "/clans"), t)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()
			g.Assert(len(result["clans"].([]interface{}))).Equal(0)
		})
	})
	g.Describe("Retrieve Clan Handler", func() {
		g.It("Should get details for clan", func() {
			player := models.PlayerFactory.MustCreate().(*models.Player)
			err = testDb.Insert(player)
			AssertNotError(g, err)

			clan := models.ClanFactory.MustCreateWithOption(map[string]interface{}{
				"GameID":   player.GameID,
				"OwnerID":  player.ID,
				"Metadata": "{\"x\": 1}",
			}).(*models.Clan)
			err = testDb.Insert(clan)
			AssertNotError(g, err)

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(player.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)), t)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			g.Assert(result["name"]).Equal(clan.Name)
			g.Assert(result["metadata"]).Equal(clan.Metadata)
			g.Assert(result["publicID"]).Equal(nil)
		})

		g.It("Should get clan members", func() {
			clan, _, _, _, err := models.GetClanWithMemberships(
				testDb, 10, "clan-details-api", "clan-details-api-clan",
			)
			AssertNotError(g, err)

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(clan.GameID, fmt.Sprintf("/clans/%s", clan.PublicID)), t)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			g.Assert(result["members"] == nil).IsFalse()
		})
	})
	g.Describe("Search Clan Handler", func() {
		g.It("Should search for a clan", func() {
			player, expectedClans, err := models.GetTestClans(
				testDb, "", "clan-apisearch-clan", 10,
			)
			g.Assert(err == nil).IsTrue()

			a := GetDefaultTestApp()
			res := Get(a, GetGameRoute(player.GameID, "clan-search?term=APISEARCH"), t)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			g.Assert(result["success"]).IsTrue()

			clans := result["clans"].([]interface{})
			g.Assert(len(clans)).Equal(10)

			for index, expectedClan := range expectedClans {
				clan := clans[index].(map[string]interface{})
				g.Assert(clan["name"]).Equal(expectedClan.Name)
			}
		})
	})
}
