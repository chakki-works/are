package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"strings"

	"github.com/urfave/cli"
)

//Gist Structure
type Gist struct {
	URL   string                 `json:"url"`
	ID    string                 `json:"id"`
	Files map[string]ContentFile `json:"files"`
}

//ContentFile Structure in the Gist
type ContentFile struct {
	FileName string `json:"filename"`
	Type     string `json:"type"`
	Content  string `json:"content"`
}

func getCandidates() map[string]string {
	const bowmeFile = ".bowme"
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Can not get user directory")
		return nil
	}
	path := usr.HomeDir + "/" + bowmeFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// get default csv from publi gist
		resp, err := http.Get("https://api.github.com/gists/55cddaa1b0c35c26cac0bace2f2b6940")
		if err != nil {
			fmt.Println("Can not get .bowme file for default setting.")
			return nil
		}

		defer resp.Body.Close()
		gist := new(Gist)
		decodeErr := json.NewDecoder(resp.Body).Decode(gist)
		if decodeErr != nil {
			fmt.Println("Can not decode the gist response.", decodeErr)
			return nil
		}
		ioutil.WriteFile(path, []byte(gist.Files["bowme.csv"].Content), 0644)
	}

	csvfile, err := os.Open(path)
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)

	candidates := map[string]string{}
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
	app.Name = "bowme"
	app.Usage = "Help you to find the commands that you remember ambiguously."
	app.Action = func(c *cli.Context) error {
		if c.NArg() == 0 {
			fmt.Println("Please input what you want to remember!")
			return nil
		}
		keyword := c.Args().Get(0)
		candidates := getCandidates()
		matched := find(keyword, candidates)
		for k, m := range matched {
			fmt.Printf("\x1b[34m%s\x1b[0m\n", k)
			fmt.Printf("  %s\n", m)
		}
		return nil
	}

	app.Run(os.Args)
}
