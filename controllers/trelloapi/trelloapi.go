package trelloapi

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/adlio/trello"
	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/config"
	"github.com/tjm/puppet-patching-automation/models"
)

var trelloClient *trello.Client

// CreateTrelloBoard Trello board via API
// Params:
// * wait (bool) - wait for trello board to be populated before returning
func CreateTrelloBoard(tb *models.TrelloBoard, waitForBoard bool, baseURL string) (err error) {
	if tb.Name == "" {
		tb.Name = "Patching: " + time.Now().Format("2006-01-02")
	}

	if tb.Description == "" {
		tb.Description = fmt.Sprintf(
			"# Patching Automation\n"+
				"* Generated: `%s`\n"+
				"* Patching Automation Tool: %s/patchRun/%v",
			time.Now().Format(time.UnixDate),
			baseURL, tb.PatchRunID)
	}

	// Background color
	if tb.Background == "" || tb.Background == "random" {
		// Choose a random color
		colors := []string{"blue", "orange", "green", "red", "purple", "pink", "lime", "sky", "grey"}
		colorIndex := rand.Intn(len(colors)) /* #nosec G404 - We do not need cryptographically strong random to pick a random color */
		tb.Background = colors[colorIndex]
	}

	board, err := createTrelloBoardAPI(tb)
	if err != nil {
		log.Error("Trello Board Creation Failed: " + err.Error())
		return
	}

	// if wait for trello board to be populated
	if waitForBoard {
		err = populateTrelloBoard(tb, board, baseURL)
		if err != nil {
			log.Error("Error populating trelloBoard: ", err)
		}
	} else {
		go func() {
			err = populateTrelloBoard(tb, board, baseURL)
			if err != nil {
				log.Error("Error populating trelloBoard: ", err)
			}
		}()
	}
	tb.Save()
	return
}

// createTrelloBoardAPI Creates a trello board with patching automation details
func createTrelloBoardAPI(tb *models.TrelloBoard) (board *trello.Board, err error) {
	board = new(trello.Board)
	board.Name = tb.Name
	board.Desc = tb.Description
	board.Prefs.Background = tb.Background
	client := getTrelloClient()
	user, err := client.GetMember("me")
	if err != nil {
		log.Error("Unable to get trello member for myself.", err)
		return
	}
	board.IDOrganization = user.IDOrganizations[0]
	board.Prefs.PermissionLevel = "org" // hard-coding for now
	board.Prefs.SelfJoin = true         // allow org members to join board
	err = getTrelloClient().CreateBoard(board, trello.Defaults())
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("Trello Board: %s <%s>", board.Name, board.ShortURL)
	// Update DB
	tb.URL = board.ShortURL
	tb.RemoteID = board.ID
	tb.Save()
	return
}

// DeleteTrelloBoard will Delete the trello board indicated using the Trello API
func DeleteTrelloBoard(tb *models.TrelloBoard) (err error) {
	var board *trello.Board
	client := getTrelloClient()
	board, err = client.GetBoard(tb.RemoteID, trello.Defaults())
	if err == nil {
		err = board.Delete(trello.Defaults())
		if err != nil {
			return
		}
	} else { // err != nil
		if trello.IsNotFound(err) {
			log.Info("Trello Board Not Found while deleting, that is OK. Error:" + err.Error())
			err = nil // 404 - Not found when deleting is not an error. Continue!
		} else {
			return
		}
	}
	err = tb.Delete(true) // return err
	return
}

// populateTrelloBoard to the trello board
func populateTrelloBoard(tb *models.TrelloBoard, board *trello.Board, baseURL string) (err error) {
	// Board Lists
	lists, err := board.GetLists(trello.Defaults())
	if err != nil {
		log.Error(err)
		return
	}

	// First List in Board (ToDo by default)
	boardList := lists[0]

	// Show progress while creating cards/checklists/items
	fmt.Printf(" -- Trello Card Progress (+=Card, #=CheckList, .=Server(Item)): ")

	// Counters for log messages
	var cardCount, checklistCount, itemCount int

	// Loop through appList
	for _, app := range models.GetApplications(tb.PatchRunID) {
		for _, env := range app.GetEnvironments() {
			// Create TrelloCard
			link := fmt.Sprintf("%s/environment/%v/components", baseURL, env.ID)
			card := &trello.Card{
				Name: fmt.Sprintf("%s [%s]", app.Name, env.Name),
				Desc: fmt.Sprintf(
					"Application: `%s`\n"+
						"Environment: `%s`\n"+
						"Patching Procedure: %s\n"+
						"Patching Automation Tool: %s",
					app.Name, env.Name, app.PatchingProcedure, link),
			}
			err := boardList.AddCard(card, trello.Defaults())
			if err != nil {
				log.Error(err)
				break
			}
			cardCount++
			fmt.Printf("+") // Trello Card Progress (new Card)
			env.TrelloCardID = card.ID
			env.TrelloCardURL = card.ShortURL
			env.Save()
			for _, component := range env.GetComponents() {
				checklist, err := getTrelloClient().CreateChecklist(card, component.Name, trello.Defaults())
				if err != nil {
					log.Error(err)
					break
				}
				component.TrelloChecklistID = checklist.ID
				component.Save()
				checklistCount++
				fmt.Printf("#") // Trello Card Progress (new Checklist)

				for _, server := range component.GetServers() {
					checked := strconv.FormatBool(server.PackageUpdates == 0)
					itemArgs := trello.Arguments{
						"pos":     "bottom",
						"checked": checked,
					}
					var item *trello.CheckItem
					if strings.Contains(server.Name, "cliqa") {
						item, err = checklist.CreateCheckItem(fmt.Sprintf("%s (%s) - Updates: %v\nssh://%s [VMName: %s]", server.Name, server.IPAddress, server.PackageUpdates, server.Name, server.VMName), itemArgs)
					} else {
						item, err = checklist.CreateCheckItem(fmt.Sprintf("%s (%s) - Updates: %v\nssh://%s", server.Name, server.IPAddress, server.PackageUpdates, server.Name), itemArgs)
					}
					if err != nil {
						log.Error(err)
						break
					}
					itemCount++
					fmt.Printf(".") // Trello Card Progress (new ChecklistItem)
					server.TrelloItemID = item.ID
				} // for component.GetServerList
			} // for env.GetComponentList
		} // for app.GetEnvironmentList
	} // for appList.GetApplications
	fmt.Printf("\n") // END of Trello Card Progress output.
	log.WithFields(log.Fields{
		"cards":      cardCount,
		"checklists": checklistCount,
		"items":      itemCount,
	}).Info("Trello Board Complete.")
	return nil
} // func CreateTrelloBoard

// getTrelloClient will return the trelloClient logged in
func getTrelloClient() *trello.Client {
	if trelloClient == nil {
		// Get Credentials
		appKey := config.GetArgs().TrelloAppKey
		token := config.GetArgs().TrelloToken
		if appKey == "" || token == "" {
			// Removed interactive prompt for WebUI
			invalidTrelloToken()
			return nil
		}
		trelloClient = trello.NewClient(appKey, token)
		// Verify client is working, as the above command just creates the object
		user, err := trelloClient.GetMember("me")
		if err != nil {
			if trello.IsPermissionDenied(err) {
				invalidTrelloToken() // Catch expired tokens
			}
			log.Error("There was a problem using the trello client: ", err)
			return nil
		}
		log.Info("Logged into Trello (token) as: " + user.FullName)
	}
	return trelloClient
}

func invalidTrelloToken() {
	appKey := config.GetArgs().TrelloAppKey
	token := config.GetArgs().TrelloToken
	if appKey == "" {
		log.Error("Trello APPKEY cannot be blank!")
		log.Error("You should set your Trello AppKey using --trelloappkey or use the TRELLO_APPKEY environment variable. (see --help for details)")
	} else if token == "" {
		log.Error("Trello TOKEN cannot be blank!")
		log.Error("You should set your Trello token using --trellotoken or use the TRELLO_TOKEN environment variable. (see --help for details)")
	} else {
		log.Error(" *** Trello AppKey/Token combination was invalid (or expired). ***")
	}
	log.Errorf("Please Visit https://trello.com/1/connect?key=%s&name=PatchingAutomation&response_type=token&scope=read,write\n\n", appKey)
}
