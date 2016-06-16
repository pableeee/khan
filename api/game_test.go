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
	"strings"
	"testing"

	"github.com/Pallinder/go-randomdata"
	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/khan/models"
)

func TestGameHandler(t *testing.T) {
	g := Goblin(t)

	//special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Create Game Handler", func() {
		g.It("Should create game", func() {
			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"publicID":                      randomdata.FullName(randomdata.RandomGender),
				"name":                          randomdata.FullName(randomdata.RandomGender),
				"metadata":                      "{\"x\": 1}",
				"minMembershipLevel":            1,
				"maxMembershipLevel":            10,
				"minLevelToAcceptApplication":   1,
				"minLevelToCreateInvitation":    1,
				"minLevelOffsetToPromoteMember": 1,
				"minLevelOffsetToDemoteMember":  1,
				"allowApplication":              true,
			}
			res := PostJSON(a, "/games", t, payload)

			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbGame, err := models.GetGameByPublicID(a.Db, payload["publicID"].(string))
			AssertNotError(g, err)
			g.Assert(dbGame.PublicID).Equal(payload["publicID"])
			g.Assert(dbGame.Name).Equal(payload["name"])
			g.Assert(dbGame.Metadata).Equal(payload["metadata"])
			g.Assert(dbGame.MinMembershipLevel).Equal(payload["minMembershipLevel"])
			g.Assert(dbGame.MaxMembershipLevel).Equal(payload["maxMembershipLevel"])
			g.Assert(dbGame.MinLevelToAcceptApplication).Equal(payload["minLevelToAcceptApplication"])
			g.Assert(dbGame.MinLevelToCreateInvitation).Equal(payload["minLevelToCreateInvitation"])
			g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(payload["minLevelOffsetToPromoteMember"])
			g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(payload["minLevelOffsetToDemoteMember"])
			g.Assert(dbGame.AllowApplication).Equal(payload["allowApplication"])
		})

		g.It("Should not create game if bad payload", func() {
			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"publicID":                      randomdata.FullName(randomdata.RandomGender),
				"name":                          randomdata.FullName(randomdata.RandomGender),
				"metadata":                      "{\"x\": 1}",
				"minMembershipLevel":            15,
				"maxMembershipLevel":            10,
				"minLevelToAcceptApplication":   1,
				"minLevelToCreateInvitation":    1,
				"minLevelOffsetToPromoteMember": 1,
				"minLevelOffsetToDemoteMember":  1,
				"allowApplication":              true,
			}
			res := PostJSON(a, "/games", t, payload)

			res.Status(422)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(len(result["reason"].([]interface{}))).Equal(3)
			g.Assert(result["reason"].([]interface{})[0].(string)).Equal("MaxMembershipLevel should be greater or equal to MinMembershipLevel")
			g.Assert(result["reason"].([]interface{})[1].(string)).Equal("MinLevelToAcceptApplication should be greater or equal to MinMembershipLevel")
			g.Assert(result["reason"].([]interface{})[2].(string)).Equal("MinLevelToCreateInvitation should be greater or equal to MinMembershipLevel")
		})

		g.It("Should not create game if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PostBody(a, "/games", t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not create game if invalid data", func() {
			a := GetDefaultTestApp()
			payload := map[string]interface{}{
				"publicID":                      "game-id-is-too-large-for-this-field-should-be-less-than-36-chars",
				"name":                          randomdata.FullName(randomdata.RandomGender),
				"metadata":                      "{\"x\": 1}",
				"minMembershipLevel":            1,
				"maxMembershipLevel":            10,
				"minLevelToAcceptApplication":   1,
				"minLevelToCreateInvitation":    1,
				"minLevelOffsetToPromoteMember": 1,
				"minLevelOffsetToDemoteMember":  1,
				"allowApplication":              true,
			}
			res := PostJSON(a, "/games", t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: value too long for type character varying(36)")
		})
	})

	g.Describe("Update Game Handler", func() {
		g.It("Should update game", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			metadata := "{\"y\": 10}"
			payload := map[string]interface{}{
				"publicID":                      game.PublicID,
				"name":                          game.Name,
				"minMembershipLevel":            game.MinMembershipLevel,
				"maxMembershipLevel":            game.MaxMembershipLevel,
				"minLevelToAcceptApplication":   game.MinLevelToAcceptApplication,
				"minLevelToCreateInvitation":    game.MinLevelToCreateInvitation,
				"minLevelOffsetToPromoteMember": game.MinLevelOffsetToPromoteMember,
				"minLevelOffsetToDemoteMember":  game.MinLevelOffsetToDemoteMember,
				"allowApplication":              true,
				"metadata":                      metadata,
			}

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, t, payload)
			res.Status(http.StatusOK)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsTrue()

			dbGame, err := models.GetGameByPublicID(a.Db, game.PublicID)
			AssertNotError(g, err)
			g.Assert(dbGame.Metadata).Equal(metadata)
			g.Assert(dbGame.PublicID).Equal(game.PublicID)
			g.Assert(dbGame.Name).Equal(game.Name)
			g.Assert(dbGame.MinMembershipLevel).Equal(game.MinMembershipLevel)
			g.Assert(dbGame.MaxMembershipLevel).Equal(game.MaxMembershipLevel)
			g.Assert(dbGame.MinLevelToAcceptApplication).Equal(game.MinLevelToAcceptApplication)
			g.Assert(dbGame.MinLevelToCreateInvitation).Equal(game.MinLevelToCreateInvitation)
			g.Assert(dbGame.MinLevelOffsetToPromoteMember).Equal(game.MinLevelOffsetToPromoteMember)
			g.Assert(dbGame.MinLevelOffsetToDemoteMember).Equal(game.MinLevelOffsetToDemoteMember)
			g.Assert(dbGame.AllowApplication).Equal(game.AllowApplication)
		})

		g.It("Should not update game if bad payload", func() {
			a := GetDefaultTestApp()
			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			payload := map[string]interface{}{
				"publicID":                      game.PublicID,
				"name":                          game.Name,
				"metadata":                      game.Metadata,
				"minMembershipLevel":            game.MaxMembershipLevel + 1,
				"maxMembershipLevel":            game.MaxMembershipLevel,
				"minLevelToAcceptApplication":   game.MinLevelToAcceptApplication,
				"minLevelToCreateInvitation":    game.MinLevelToCreateInvitation,
				"minLevelOffsetToPromoteMember": game.MinLevelOffsetToPromoteMember,
				"minLevelOffsetToDemoteMember":  game.MinLevelOffsetToDemoteMember,
				"allowApplication":              game.AllowApplication,
			}
			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, t, payload)

			res.Status(422)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(len(result["reason"].([]interface{}))).Equal(3)
			g.Assert(result["reason"].([]interface{})[0].(string)).Equal("MaxMembershipLevel should be greater or equal to MinMembershipLevel")
			g.Assert(result["reason"].([]interface{})[1].(string)).Equal("MinLevelToAcceptApplication should be greater or equal to MinMembershipLevel")
			g.Assert(result["reason"].([]interface{})[2].(string)).Equal("MinLevelToCreateInvitation should be greater or equal to MinMembershipLevel")
		})

		g.It("Should not update game if invalid payload", func() {
			a := GetDefaultTestApp()
			res := PutBody(a, "/games/game-id", t, "invalid")

			res.Status(http.StatusBadRequest)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(strings.Contains(result["reason"].(string), "While trying to read JSON")).IsTrue()
		})

		g.It("Should not update game if invalid data", func() {
			a := GetDefaultTestApp()

			game := models.GameFactory.MustCreate().(*models.Game)
			err := a.Db.Insert(game)
			AssertNotError(g, err)

			metadata := ""

			payload := map[string]interface{}{
				"publicID":                      game.PublicID,
				"name":                          game.Name,
				"minMembershipLevel":            game.MinMembershipLevel,
				"maxMembershipLevel":            game.MaxMembershipLevel,
				"minLevelToAcceptApplication":   game.MinLevelToAcceptApplication,
				"minLevelToCreateInvitation":    game.MinLevelToCreateInvitation,
				"minLevelOffsetToPromoteMember": game.MinLevelOffsetToPromoteMember,
				"minLevelOffsetToDemoteMember":  game.MinLevelOffsetToDemoteMember,
				"allowApplication":              true,
				"metadata":                      metadata,
			}

			route := fmt.Sprintf("/games/%s", game.PublicID)
			res := PutJSON(a, route, t, payload)

			res.Status(http.StatusInternalServerError)
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			g.Assert(result["success"]).IsFalse()
			g.Assert(result["reason"]).Equal("pq: invalid input syntax for type json")
		})
	})
}