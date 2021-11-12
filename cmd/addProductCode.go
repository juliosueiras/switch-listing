package cmd

import (
  "os"

  "fmt"
	"github.com/spf13/cobra"
	"github.com/cheggaaa/pb/v3"
  "regexp"
  "strings"
  "log"
  "net/http"
  "io"
  "encoding/csv"
	"github.com/manifoldco/promptui"
  "github.com/sahilm/fuzzy"
  "sort"
)

var addProductCodeCmd = &cobra.Command{
	Use:   "add-product-code",
	Run: func(cmd *cobra.Command, args []string) {
    csvFileName, _ := cmd.PersistentFlags().GetString("game-input-file")
    codesFileName, _ := cmd.PersistentFlags().GetString("code-region-input-file")
    outputFileName, _ := cmd.PersistentFlags().GetString("output-file")
    csvFile, _ := readData(csvFileName)
    codesFile, err := readDataMod(codesFileName)

    if err != nil {
      log.Fatal(err)
    }

    codesMap := make(map[string]ProductCode)

    for _, record := range codesFile {
      codesMap[record[1]] = ProductCode{
        Name: record[1],
        Title: record[0],
        Region: record[2],
      }
    }

    client := &http.Client{}
    req, err := http.NewRequest("GET", "https://www.gametdb.com/switchtdb.txt", nil)

    if err != nil {
    }

    req.Header.Set("User-Agent", "Switch-Listing/3.0")

    resp, err := client.Do(req)

    if err != nil {
    }

    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)

    reg, err := regexp.Compile("[^a-zA-Z0-9]+")
    if err != nil {
      log.Fatal(err)
    }

    switchdb := string(body)

    templates := &promptui.SelectTemplates{
      Label:    "{{ . }}?",
      Active:   "{{ .Title | cyan }} | {{ .Name | cyan }} ({{ .Region | red }})",
      Inactive: "{{ .Title | cyan }} | {{ .Name | cyan }} ({{ .Region | blue }})",
      Selected: "{{ .Title | red | cyan }} | {{ .Name | red | cyan }}",
    }

    switchdbLines := strings.Split(switchdb, "\n")

    titles := make(map[string][]string, 0)
    titlesString := make([]string, 0)

    for _, i := range switchdbLines[1:] {
      line := strings.Split(i, "=")
      processedString := strings.ToLower(reg.ReplaceAllString(strings.Join(line[1:], "="), ""))
      titles[processedString] = append(titles[processedString], strings.TrimSpace(line[0]))
      titlesString = append(titlesString, processedString)
    }


    bar := pb.StartNew(len(csvFile))
    finalMatches := make([]string, 0)
    for _, record := range csvFile {
      processedString := strings.ToLower(reg.ReplaceAllString(record[0], ""))
      matches := fuzzy.Find(processedString, titlesString)
      sort.Sort(matches)
      matchText := "Not Found"
      if matches.Len() != 0 {
        matchText = titles[matches[0].Str][0]

        if len(titles[matches[0].Str]) > 1 {
          codes := make([]ProductCode, 0)
          for _, v := range titles[matches[0].Str] {
            codes = append(codes, codesMap[v])
          }

          codes = append(codes, ProductCode{
            Name: "",
            Title: "",
            Region: "Enter Custom ID",
          })

          prompt := promptui.Select{
            Label: "Select Product Code",
            Templates: templates,
            Items: codes,
            HideSelected: true,
          }

          i, _, err := prompt.Run()

          if err != nil {
            os.Exit(1)
          }

          if i == len(codes)-1 {
            prompt := promptui.Prompt{
              Label: fmt.Sprintf("Enter Custom ID for %s", codes[0].Title),
            }

            result, err := prompt.Run()

            if err != nil {
              log.Fatalf("Prompt failed %v\n", err)
              os.Exit(1)
            }

            matchText = result
          } else {
            matchText = codes[i].Name
          }
        }
      }

      finalMatches = append(finalMatches, matchText)
      bar.Increment()
    }

    err = os.WriteFile(outputFileName, []byte(strings.Join(finalMatches, "\n")), 0644)

    if err != nil {
      log.Fatalf(err.Error())
    }

    bar.Finish()
  },
}

func init() {
  addProductCodeCmd.PersistentFlags().String("game-input-file", "export_games.csv", "Game Input CSV File")
  addProductCodeCmd.PersistentFlags().String("code-region-input-file", "code_region.txt", "Code Region File(generated using script)")
  addProductCodeCmd.PersistentFlags().String("output-file", "finished.txt", "Output file")
  rootCmd.AddCommand(addProductCodeCmd)
}

type ProductCode struct {
  Name string
  Title string
  Region string
}

func readData(fileName string) ([][]string, error) {

  f, err := os.Open(fileName)

  if err != nil {
    return [][]string{}, err
  }

  defer f.Close()

  r := csv.NewReader(f)

  // skip first line
  if _, err := r.Read(); err != nil {
    return [][]string{}, err
  }

  records, err := r.ReadAll()

  if err != nil {
    return [][]string{}, err
  }

  return records, nil
}

func readDataMod(fileName string) ([][]string, error) {

  f, err := os.Open(fileName)

  if err != nil {
    return [][]string{}, err
  }

  defer f.Close()

  r := csv.NewReader(f)
  r.Comma = '|'
  r.LazyQuotes = true

  records, err := r.ReadAll()

  if err != nil {
    return [][]string{}, err
  }

  return records, nil
}
