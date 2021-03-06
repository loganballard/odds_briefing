package oddsapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/loganballard/odds_briefing/config"
	"github.com/loganballard/odds_briefing/logger"
)

const baseAPIURL string = "https://api.the-odds-api.com"
const region string = "us" // only bet on USA!

// SHARED FUNCTIONS

func getOddsAPIKey() string {
	var credFile config.Credentials
	credFile.LoadCredentials()
	return credFile.OddsAPIKey
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func h2hToAmericanOdds(h2hOdds string) int {
	floatOdds, err := strconv.ParseFloat(h2hOdds, 64)
	if err != nil {
		logger.ErrorHelper(err)
	}
	floatOdds = round(floatOdds, 0.01)
	if floatOdds >= 2.0 {
		return int((floatOdds - 1) * 100)
	}
	return int(-100 / (floatOdds - 1))
}

// MakeAPIRequest sends an API get Request to a specific endpoint and fails if not 2xx error code
// TODO - rewrite as http client rather than endpoint
func MakeAPIRequest(endpoint string) []byte {
	finalURL := baseAPIURL + endpoint

	resp, err := http.Get(finalURL)
	if err != nil {
		logger.ErrorHelper(err)
	}
	if (resp.StatusCode >= 200) && (resp.StatusCode >= 300) {
		logger.Warn(fmt.Sprintf("non 2xx error code for: %s", finalURL))
		logger.ErrorHelper(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorHelper(err)
	}

	return body
}

// END SHARED FUNCTIONS

// ACTIVE SPORTS API CALL FUNCTIONS

func processActiveSportsResponse(jsonResponseBody []byte) ActiveSportsResponse {
	var decodedActiveSportsResp ActiveSportsResponse
	err := json.Unmarshal(jsonResponseBody, &decodedActiveSportsResp)

	if err != nil {
		logger.ErrorHelper(err)
	}

	return decodedActiveSportsResp
}

func getListOfSportsFromActiveResp(decodedActiveSportsResp *ActiveSportsResponse) []string {
	var listOfSports []string

	for _, entry := range (*decodedActiveSportsResp).Data {
		listOfSports = append(listOfSports, entry.Key)
	}

	return listOfSports
}

// END ACTIVE SPORTS API CALL FUNCTIONS

// NFL TOTALS API CALL FUNCTIONS

func convertTotalsPointsToStringAndFloat(totals *totalsOddsResponse) {
	for i, entry := range totals.Games {
		for j, site := range entry.Sites {
			for k, pts := range site.Odds.Totals.Points {
				totals.Games[i].Sites[j].Odds.Totals.PointsStr = append(totals.Games[i].Sites[j].Odds.Totals.PointsStr, fmt.Sprintf("%v", pts))
				flt, err := strconv.ParseFloat(totals.Games[i].Sites[j].Odds.Totals.PointsStr[k], 32)
				if err != nil {
					logger.ErrorHelper(err)
				}
				totals.Games[i].Sites[j].Odds.Totals.PointsFloat = append(totals.Games[i].Sites[j].Odds.Totals.PointsFloat, flt)
			}
		}
	}
}

func processNflTotalsResponse(jsonResponseBody []byte) totalsOddsResponse {
	var decodedNflTotalsResp totalsOddsResponse
	err := json.Unmarshal(jsonResponseBody, &decodedNflTotalsResp)
	if err != nil {
		logger.ErrorHelper(err)
	}

	convertTotalsPointsToStringAndFloat(&decodedNflTotalsResp)

	logger.Info(fmt.Sprintf("NFL Totals API Response: %v\n", decodedNflTotalsResp))

	return decodedNflTotalsResp
}

func makeTwoTeamNamesIntoOne(teamNames []string) string {
	var bothNames string = ""
	for _, name := range teamNames {
		bothNames = bothNames + name + " "
	}
	return bothNames
}

func makeAdjustedOverUnder(sites []oddsTotalsSiteEntry) (float64, float64) {
	var totalOverUnder float64 = 0
	for _, site := range sites {
		totalOverUnder += site.Odds.Totals.PointsFloat[0]
	}
	adjustedOverUnder := totalOverUnder / float64(len(sites))
	return adjustedOverUnder, adjustedOverUnder
}

// TODO - implement
func makeAdjustedOverUnderOdds() (int, int) {
	return -110, -110
}

func formatNflTotalsResp(totalsResp totalsOddsResponse) FormattedTotalsOdds {
	var totalOdds FormattedTotalsOdds
	totalOdds.OddsType = "Totals"
	totalOdds.Sport = "NFL"
	for _, game := range totalsResp.Games {
		var gameTotals TotalOdds
		gameTime := time.Unix(game.Gametime, 0)
		gameTotals.Gametime = gameTime.Local()
		gameTotals.Teams = makeTwoTeamNamesIntoOne(game.Teams)
		gameTotals.Over, gameTotals.Under = makeAdjustedOverUnder(game.Sites)
		gameTotals.OverOdds, gameTotals.UnderOdds = makeAdjustedOverUnderOdds()
		totalOdds.Odds = append(totalOdds.Odds, gameTotals)
	}
	return totalOdds
}

func formatNflTotalsMessageString(odds TotalOdds) string {
	var msg strings.Builder
	tm := time.Now()
	zone, _ := tm.Zone()

	msg.WriteString(fmt.Sprintf("%s\n", odds.Teams))
	msg.WriteString(fmt.Sprintf("%02d/%02d at %2d:%02d %s\n", odds.Gametime.Month(), odds.Gametime.Day(), odds.Gametime.Hour(), odds.Gametime.Minute(), zone))
	msg.WriteString(fmt.Sprintf("Adjusted Over Under: %0.1f\n", odds.Over))
	msg.WriteString(fmt.Sprintf("Over odds: %d\nUnder odds: %d\n", odds.OverOdds, odds.UnderOdds))
	msg.WriteString("Best of Luck!")

	return msg.String()
}
