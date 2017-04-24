package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"runtime"
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

const defaultLoadURL = "https://api.github.com/gists/55cddaa1b0c35c26cac0bace2f2b6940"
const bowmeFile = ".bowme"

func getBowmePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.New("Can not get user directory")
	}
	path := usr.HomeDir + "/" + bowmeFile
	return path, nil
}

func writeBowmeFile(url string, path string) error {
	target := path
	if bPath, err := getBowmePath(); path == "" && bPath != "" {
		if err == nil {
			target = bPath
		} else {
			return errors.New("Can not set default path (home directory)")
		}
	}

	// get default csv from publi gist
	resp, err := http.Get(defaultLoadURL)
	if err != nil {
		return errors.New("Can not get .bowme file for default setting")
	}

	defer resp.Body.Close()
	gist := new(Gist)
	decodeErr := json.NewDecoder(resp.Body).Decode(gist)
	if decodeErr != nil {
		emsg := fmt.Sprintf("Can not decode the gist response. %s", decodeErr)
		return errors.New(emsg)
	}
	ioutil.WriteFile(target, []byte(gist.Files["bowme.csv"].Content), 0644)
	return nil
}

func appendBowmeFile(command string, index string, path string) error {
	target := path
	if bPath, err := getBowmePath(); path == "" && bPath != "" {
		if err == nil {
			target = bPath
		} else {
			return errors.New("Can not find the bowme file")
		}
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return errors.New("bowme file does not exist")
	}

	//write to file
	f, err := os.OpenFile(target, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("Can not open the file")
	}
	defer f.Close()

	content := strings.Join([]string{index, command}, ",") + "\n"
	if _, err = f.WriteString(content); err != nil {
		return errors.New("Can not write to the file")
	}

	return nil

}

func getCandidates() map[string]string {
	bowmePath, err := getBowmePath()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if _, err := os.Stat(bowmePath); os.IsNotExist(err) {
		err := writeBowmeFile(defaultLoadURL, bowmePath)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}

	csvfile, err := os.Open(bowmePath)
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)

	candidates := map[string]string{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err == csv.ErrFieldCount {
				break // ignore
			} else {
				fmt.Println("Can not read the csv: ", err)
				break
			}
		}
		key := strings.TrimSpace(record[0])
		command := strings.TrimSpace(record[1])
		candidates[key] = command
	}
	return candidates
}

func find(keyword string, candidates map[string]string) map[string]string {
	query := strings.Join(strings.Split(keyword, " "), "|")
	r := regexp.MustCompile("(?i)" + query)

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

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "reload, r",
			Value: "",
			Usage: "Read the .bowme file from Gist URL (Gist should include the bowme.csv file)",
		},
		cli.StringFlag{
			Name:  "index, i",
			Value: "",
			Usage: "The index for command that is added to bowme file",
		},
	}
	app.Action = func(c *cli.Context) error {
		reload := c.String("reload")
		index := c.String("index")
		if reload != "" {
			err := writeBowmeFile(reload, "")
			if err == nil {
				fmt.Println(fmt.Sprintf("Read .bowme file from %s !", reload))
			} else {
				fmt.Println(err)
			}
		} else if index != "" {
			if c.NArg() == 0 {
				fmt.Println("Please input what command you want to add.")
				return nil
			}
			cmd := strings.Join(c.Args(), " ")
			err := appendBowmeFile(cmd, index, "")
			if err == nil {
				fmt.Println(fmt.Sprintf("Append to bowme file: %s:%s", index, cmd))
			} else {
				fmt.Println(err)
			}
		} else {
			if c.NArg() == 0 {
				fmt.Println("Please input what you want to remember!")
				return nil
			}
			keyword := c.Args().Get(0)
			candidates := getCandidates()
			matched := find(keyword, candidates)
			for k, m := range matched {
				if runtime.GOOS == "windows" {
					fmt.Printf("[%s]", k)
				} else {
					fmt.Printf("\x1b[34m%s\x1b[0m\n", k)
				}
				fmt.Printf("  %s\n", m)
			}
		}
		return nil
	}

	app.Run(os.Args)
}
