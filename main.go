package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli"
)

type Gist struct {
	URL   string                 `json:"url"`
	ID    string                 `json:"id"`
	Files map[string]ContentFile `json:"files"`
}

type ContentFile struct {
	FileName string `json:"filename"`
	Type     string `json:"type"`
	Content  string `json:"content"`
}

func getCandidates() map[string]string {
	gistID := os.Getenv("ARE_GIST_ID")
	if gistID == "" {
		gistID = "55cddaa1b0c35c26cac0bace2f2b6940"
	}
	resp, err := http.Get("https://api.github.com/gists/" + gistID)
	if err != nil {
		fmt.Println("Can not access to the gist. Please check ARE_GIST_ID environmental variable or network connection.")
		return nil
	}
	defer resp.Body.Close()
	gist := new(Gist)
	decodeErr := json.NewDecoder(resp.Body).Decode(gist)
	if decodeErr != nil {
		fmt.Println("Can not decode the gist response.", decodeErr)
		return nil
	}

	candidates := map[string]string{}
	reader := csv.NewReader(strings.NewReader(gist.Files["are.csv"].Content))
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Can not read the csv: ", err)
			break
		}
		key := strings.TrimSpace(record[0])
		command := strings.TrimSpace(record[1])
		candidates[key] = command
	}
	return candidates
}

func find(keyword string, candidates map[string]string) map[string]string {
	query := strings.Join(strings.Split(keyword, " "), "|")
	r := regexp.MustCompile(query)

	matched := map[string]string{}
	for k, c := range candidates {
		if r.MatchString(k) {
			matched[k] = c
		}
	}
	return matched
}

func main() {
	app := cli.NewApp()
	app.Name = "are"
	app.Usage = "type what you want to remember."
	app.Action = func(c *cli.Context) error {
		if c.NArg() == 0 {
			fmt.Println("Please input what you want to remember!")
			return nil
		}
		are := c.Args().Get(0)
		candidates := getCandidates()
		matched := find(are, candidates)
		for k, m := range matched {
			fmt.Printf("\x1b[34m%s\x1b[0m\n", k)
			fmt.Printf("  %s\n", m)
		}
		return nil
	}

	app.Run(os.Args)
}
