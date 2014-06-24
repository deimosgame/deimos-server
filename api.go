package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type APIRequest struct {
	Request  *http.Request
	Response *http.Response
	Player   *Player
	Callback func(*APIRequest)

	AchievementId int
}

// WebProcess is intended to run as a goroutine. It manages all the requests to
// the web server
func APIProcess() {
	for {
		r, ok := <-APIInput
		if !ok {
			return
		}
		res, err := http.DefaultClient.Do(r.Request)
		if err != nil {
			if !apiServerLost {
				apiServerLost = true
				log.Warn("Lost connection to the master server!")
			}
			continue
		}
		apiServerLost = false
		if res.StatusCode != 200 {
			log.Warn("Error while contacting the master server (" +
				strconv.Itoa(res.StatusCode) + ")")
		}
		r.Response = res
		r.Callback(r)
	}
}

// CheckUnlockedAchievements initiates the request for the list of achievements
// a player unlocked
func CheckUnlockedAchivements(p *Player) {
	req, _ := http.NewRequest("GET", APIServer+"/unlocked-achievements/"+
		p.Account, nil)
	APIInput <- &APIRequest{
		Request:  req,
		Player:   p,
		Callback: CheckUnlockedAchivementsCallback,
	}
}

// CheckUnlockedAchivementsCallback saves the list of achievements a player
// unlocked into the Player struct
func CheckUnlockedAchivementsCallback(apiReq *APIRequest) {
	body, err := ioutil.ReadAll(apiReq.Response.Body)
	if err != nil {
		return
	}
	type Response struct {
		Success bool
		List    []int
	}
	var response Response
	json.Unmarshal(body, &response)
	if !response.Success {
		log.Warn("Error while checking achievements of " +
			apiReq.Player.Account)
		return
	}
	apiReq.Player.Achievements = response.List
}

// UnlockAchievement initiates the unlocking of an achievement for a player
func UnlockAchievement(p *Player, AchievementId int) {
	for _, currentAchievement := range p.Achievements {
		if currentAchievement == AchievementId {
			return
		}
	}
	req, _ := http.NewRequest("GET", APIServer+"/unlock-achievements/"+
		p.Account+"/"+strconv.Itoa(AchievementId), nil)
	APIInput <- &APIRequest{
		Request:       req,
		Player:        p,
		Callback:      UnlockAchievementCallback,
		AchievementId: AchievementId,
	}
}

// UnlockAchievementCallback processes the response of the API for an
// achievement request
func UnlockAchievementCallback(apiReq *APIRequest) {
	body, err := ioutil.ReadAll(apiReq.Response.Body)
	if err != nil {
		return
	}
	type Response struct {
		Success bool
		Message string
	}
	var response Response
	json.Unmarshal(body, &response)
	if !response.Success {
		// Achievement 1 is unlocked on connect when the list may be not loaded
		if len(apiReq.Player.Achievements) > 0 && apiReq.AchievementId != 1 {
			log.Warn("Error while unlocking achievement " +
				strconv.Itoa(apiReq.AchievementId) + " for " +
				apiReq.Player.Account)
		}
		return
	}
	if len(apiReq.Player.Achievements) > 0 {
		apiReq.Player.Achievements = append(apiReq.Player.Achievements,
			apiReq.AchievementId)
	}
	SendMessage(fmt.Sprintf("> %s has unlocked the achievement %s!",
		apiReq.Player.Name, response.Message))
}

func OnPlayerKill(killed, killer *Player) {
	if killer.Equals(killed) {
		// Achivement: A Special Kind of Stupid
		UnlockAchievement(killed, 8)

		if killer.Score > 0 {
			killer.Score--
		}
		return
	}

	// Achivement: Rough Beginnings
	UnlockAchievement(killed, 6)

	// Achievement: Let the fun begin
	UnlockAchievement(killer, 2)

	killer.Victims++
	killer.Score++

	if killer.Victims == 5 {
		// Achievement: Getting used to it
		UnlockAchievement(killer, 3)
	}

	if killer.CurrentWeapon == 0 {
		// Achievement: Sharp knife
		UnlockAchievement(killer, 5)
	}

	switch killer.CurrentStreak {
	case 5:
		// Achievement: Kill streak
		UnlockAchievement(killer, 9)
	case 10:
		// Achievement: On a Rampage
		UnlockAchievement(killer, 10)
	case 20:
		// Achievement: Domination
		UnlockAchievement(killer, 11)
	}
}
