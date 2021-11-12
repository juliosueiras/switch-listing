package cmd

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/juliosueiras/switch-listing/utils"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"strings"
)

var listNSCollectorsCmd = &cobra.Command{
	Use: "list-nscollectors-games",
	Run: func(cmd *cobra.Command, args []string) {
		res, err := http.Get("https://docs.google.com/spreadsheets/d/1FNyvbbU64Pb9lheg28gC_5fMalIYJ0aD763T7M1QqF0/gviz/tq?tqx=out:csv")

		clientsString, _ := ioutil.ReadAll(res.Body)

		headerString := "gametitle,release,usadate,jpndate,eurdate,ausdate,usacart,jpncart,eurcart,auscart,english,notes\n"

		clientsString = []byte(headerString + strings.Join(strings.Split(string(clientsString), "\n")[1:], "\n"))

		codeFile, _ := cmd.PersistentFlags().GetString("code-file")
		file, _ := ioutil.ReadFile(codeFile)
		if err != nil {
			panic(err)
		}

		gameMap := make(map[string]*utils.NSCollectorSheetItem)

		ownedCodes := make(map[string]int)

		for i, v := range strings.Split(string(file), "\n") {
			if v == "Not Found" {
				continue
			}

			ownedCodes[v] = i
		}

		clients := []*utils.NSCollectorSheetItem{}

		if err := gocsv.UnmarshalBytes(clientsString, &clients); err != nil { // Load clients from file
			panic(err)
		}

		for _, client := range clients {
			gameMap[client.GameTitle] = client
		}

		test := tview.NewList()

		test1 := tview.NewTextView().SetDynamicColors(true)

		test1.SetBorder(true).SetTitle("Game Info")

		test.SetBorder(true)

		testText := func(index int, text string, secondary string, shortcut rune) {
			client := gameMap[text]
			codes := checkCartID(client)

			finalText := fmt.Sprintf(`
[red]Game Title:[white] %s
[red]Notes:[white] %s
[red]English on Cart:[white] %s

[yellow]Cart IDs[white]
[red]USA:[white] %s
[red]EU:[white] %s
[red]JPN:[white] %s
[red]AUS:[white] %s

[yellow]Release Dates[white]
[red]USA:[white] %s
[red]EU:[white] %s
[red]JPN:[white] %s
[red]AUS:[white] %s
`, client.GameTitle, client.Notes, client.EnglishOnCart, codes["USA"], codes["EU"], codes["JPN"], codes["AUS"], client.USADate, client.EUDate, client.JPNDate, client.AUSDate)
			test1.SetText(finalText)
		}

		ownedCount := 0
		for _, client := range clients {

			shortcut := '-'
			if checkOwned(client, ownedCodes) {
				ownedCount++
				shortcut = 'O'
			}

			test.
				AddItem(client.GameTitle, "", shortcut, nil).
				ShowSecondaryText(false).
				SetChangedFunc(testText)
		}

		test.SetTitle(fmt.Sprintf("Game Title (Owned %d/%d)", ownedCount, len(clients)))

		app := tview.NewApplication()

		usaCheckbox := tview.NewCheckbox()
		usaCheckbox.SetLabel("US: ")

		euCheckbox := tview.NewCheckbox()
		euCheckbox.SetLabel("EU: ")

		jpnCheckbox := tview.NewCheckbox()
		jpnCheckbox.SetLabel("JPN: ")

		ausCheckbox := tview.NewCheckbox()
		ausCheckbox.SetLabel("AUS: ")

		idOnlyCheckbox := tview.NewCheckbox()
		idOnlyCheckbox.SetLabel("Show Games with Cart ID Only: ")
		idOnlyCheckbox.SetChangedFunc(func(checked bool) {
			if checked {
				test.Clear()
				for _, client := range clients {
					shortcut := '-'
					if checkOwned(client, ownedCodes) {
						shortcut = 'O'
					}

					if codes := checkCartID(client); len(codes) != 0 {
						test.
							AddItem(client.GameTitle, "", shortcut, nil).
							ShowSecondaryText(false).
							SetChangedFunc(testText)
					}
				}
			}
		})

		checkboxRegion := func(region string) func(checked bool) {
			return func(checked bool) {
				if checked {
					test.Clear()
					for _, client := range clients {
						date := ""

						idOnlyFlag := idOnlyCheckbox.IsChecked()
						cartIdExist := false
						codes := checkCartID(client)

						switch region {
						case "USA":
							jpnCheckbox.SetChecked(false)
							euCheckbox.SetChecked(false)
							ausCheckbox.SetChecked(false)

							if idOnlyFlag {
								cartIdExist = codes["USA"] != ""
							}

							date = client.USADate
						case "EU":
							jpnCheckbox.SetChecked(false)
							usaCheckbox.SetChecked(false)
							ausCheckbox.SetChecked(false)

							if idOnlyFlag {
								cartIdExist = codes["EU"] != ""
							}
							date = client.EUDate
						case "JPN":
							euCheckbox.SetChecked(false)
							usaCheckbox.SetChecked(false)
							ausCheckbox.SetChecked(false)
							if idOnlyFlag {
								cartIdExist = codes["JPN"] != ""
							}
							date = client.JPNDate
						case "AUS":
							euCheckbox.SetChecked(false)
							usaCheckbox.SetChecked(false)
							jpnCheckbox.SetChecked(false)
							if idOnlyFlag {
								cartIdExist = codes["AUS"] != ""
							}
							date = client.AUSDate
						}
						res := strings.Split(date, "/")

						shortcut := '-'
						if checkOwned(client, ownedCodes) {
							shortcut = 'O'
						}

						if len(res) == 3 {
							if !idOnlyCheckbox.IsChecked() || cartIdExist {
								test.
									AddItem(client.GameTitle, "", shortcut, nil).
									ShowSecondaryText(false).
									SetChangedFunc(testText)
							}
						}
					}
				}
			}
		}

		usaCheckbox.SetChangedFunc(checkboxRegion("USA"))
		ausCheckbox.SetChangedFunc(checkboxRegion("AUS"))
		jpnCheckbox.SetChangedFunc(checkboxRegion("JPN"))
		euCheckbox.SetChangedFunc(checkboxRegion("EU"))

		newFlex := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewTextView().SetText("Regions:"), 1, 1, false).
			AddItem(usaCheckbox, 1, 1, false).
			AddItem(euCheckbox, 1, 1, false).
			AddItem(jpnCheckbox, 1, 1, false).
			AddItem(ausCheckbox, 1, 1, false).
			AddItem(tview.NewTextView().SetText("Other:"), 1, 1, false).
			AddItem(idOnlyCheckbox, 1, 1, false)
		newFlex.SetTitle("Game Filter Selection").SetBorder(true)
		//usaCheckbox.SetBorder(true)
		flex := tview.NewFlex().
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(test, 0, 2, true).
				AddItem(newFlex, 0, 1, false), 0, 1, true).
			AddItem(test1, 0, 1, false)

		if err := app.SetRoot(flex, true).EnableMouse(true).SetFocus(flex).Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	listNSCollectorsCmd.PersistentFlags().String("code-file", "code.txt", "Cart Code List File")
	rootCmd.AddCommand(listNSCollectorsCmd)
}

func checkOwned(client *utils.NSCollectorSheetItem, ownedCodes map[string]int) bool {
	res := strings.Split(client.USACartID, "-")

	usaCartID := "None"

	if len(res) >= 2 {
		usaCartID = res[2]
	}

	res = strings.Split(client.EUCartID, "-")

	euCartID := "None"

	if len(res) >= 2 {
		euCartID = res[2]
	}

	res = strings.Split(client.JPNCartID, "-")

	jpnCartID := "None"

	if len(res) >= 2 {
		jpnCartID = res[2]
	}

	res = strings.Split(client.AUSCartID, "-")

	ausCartID := "None"

	if len(res) >= 2 {
		ausCartID = res[2]
	}

	if _, exists := ownedCodes[usaCartID]; exists {
		return true
	} else if _, exists := ownedCodes[jpnCartID]; exists {
		return true
	} else if _, exists := ownedCodes[euCartID]; exists {
		return true
	} else if _, exists := ownedCodes[ausCartID]; exists {
		return true
	}

	return false
}

func checkCartID(client *utils.NSCollectorSheetItem) map[string]string {
	result := make(map[string]string)
	res := strings.Split(client.USACartID, "-")

	if len(res) >= 2 {
		result["USA"] = res[2]
	}

	res = strings.Split(client.JPNCartID, "-")

	if len(res) >= 2 {
		result["JPN"] = res[2]
	}

	res = strings.Split(client.EUCartID, "-")

	if len(res) >= 2 {
		result["EU"] = res[2]
	}

	res = strings.Split(client.AUSCartID, "-")

	if len(res) >= 2 {
		result["AUS"] = res[2]
	}

	return result
}
