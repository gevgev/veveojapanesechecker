package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/readline.v1"
)

const (
	xpid   = "pkg00@PASSPORT3496.Rovi"
	custid = "passport"
	rpr    = "10"
	//	ectJson = "5"
	baseUrl = "http://roviapi.veveo.net/search"
)

var params map[string]string

func init() {
	params = make(map[string]string)

	params["XPID"] = xpid
	params["custid"] = custid
	params["RPR"] = rpr
	//params["ECT"] = ectJson
	//map["W"] = ""
}

// get keyboard input for the "name"
func prompt(name, historyFileName string) string {
	prmpt := fmt.Sprintf("%v: ", name)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      prmpt,
		HistoryFile: "/tmp/" + historyFileName + ".tmp",
	})

	if err != nil {
		panic(err)
	}
	defer rl.Close()

	line, err := rl.Readline()
	if err != nil { // io.EOF
		return "quit"
	}
	return line
}

// GGuide SI BSD XML struct here -------

// Top node = ListingGroup
type Query struct {
	GemstarChannelID string
	ListingsList     []Listing `xml:"Listing"`
}

type Listing struct {
	Kind                 string
	ProgramId            string                   `xml:"ProgramID"`
	ShortEventDescriptor ShortEventDescriptorType `xml:"short_event_descriptor"`
}

type ShortEventDescriptorType struct {
	Title string
}

func (l Listing) String() string {
	return fmt.Sprintf("Kind:[%s] - ID[%s] - Title[%s]", l.Kind, l.ProgramId, l.ShortEventDescriptor)
}

func (s ShortEventDescriptorType) String() string {
	return fmt.Sprintf("%s", s.Title)
}

// GGuide SI BSD XML struct ends -------

// Veveo Response XML Structure here -------

// Top node = VtvRsps
type VeveoResponse struct {
	RC RCType
}

type RCType struct {
	CNs CNsType
}

type CNsType struct {
	N  int `xml:"n"`
	CI []CIType
}

type CIType struct {
	T      string
	G      string     `xml:"G",omitempty`
	TYP    string     `xml:"TYP",omitempty`
	ZL     string     `xml:"ZL",omitempty`
	Mtinfo MTinfoType `xml:"mtinfo",omitempty`
}

type MTinfoType struct {
	Mt   string `xml:"mt"`
	Mtfn string `xml:"mtfn"`
}

func (ci CIType) String() string {
	return fmt.Sprintf("T [%s] - G [%s] - TYP [%s] - ZL [%s] mtinfo[%s]", ci.T, ci.G, ci.TYP, ci.ZL, ci.Mtinfo)
}

func (mtinfo MTinfoType) String() string {
	return fmt.Sprintf("mt [%s] - mtfn [%s]", mtinfo.Mt, mtinfo.Mtfn)
}

// Veveo Response XML Structure ends -------

// Search ------------
func searchTerm() (searchString string) {
	fmt.Println("Enter search term (empty string to end):")
	//fmt.Scanln(&searchString)
	searchString = prompt("Search Term", "searchTerm")
	return searchString
}

func getUrl(searchTerm string) string {
	//url := fmt.Sprintf(baseUrl, xpid, custid, rpr, ectJson, searchTerm)
	var Url *url.URL
	Url, err := url.Parse(baseUrl)

	if err != nil {
		panic("Wrong base URL: " + baseUrl)
	}

	parameters := url.Values{}
	for k, v := range params {
		parameters.Add(k, v)
	}

	parameters.Add("W", searchTerm)
	Url.RawQuery = parameters.Encode()

	return Url.String()
}

func search(searchterm string) string {

	url := getUrl(searchterm)
	response, err := http.Get(url)

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		fmt.Println("--------------------begin---------------------------")
		fmt.Println("Search term: ", searchterm)
		fmt.Println("RQ:" + url)
		fmt.Println("--------------------begin---------------------------")
		fmt.Printf("%s\n", string(contents))
		fmt.Println("---------------------end----------------------------")

		return string(contents)
	}
	return ""
}

// Get the list of all XML files in the provided folder
func getXmlSIFiles(searchDir string) []string {
	fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, _ error) error {
		// fmt.Println("adding to the list: ", path)
		fileList = append(fileList, path)
		return nil
	})
	if err != nil {
		fmt.Println("Error getting files list: ", err)
	}
	return fileList
}

// Loads the GGuide SI XML file, gets the list of titles
func loadXmlFile(filename string) []string {
	titleList := []string{}

	xmlFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return titleList
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	var q Query
	xml.Unmarshal(b, &q)

	fmt.Println("Channel ID: ", q.GemstarChannelID)
	for _, listing := range q.ListingsList {
		fmt.Printf("\t%s\n", listing)
		if listing.ShortEventDescriptor.Title != "" {
			titleList = append(titleList, listing.ShortEventDescriptor.Title)
		}
	}
	return titleList
}

// Analise if we get any search results back
func analyseResults(title, responseXmlString string) {
	var responseXML VeveoResponse

	fmt.Println("Loading/unmarshalling Veveo response XML...")
	err := xml.Unmarshal([]byte(responseXmlString), &responseXML)
	if err != nil {
		fmt.Printf("Error unmarshalling Veveo response for: [%s], error: [%s]\n", title, err)
	}

	if responseXML.RC.CNs.N == 1 {
		if responseXML.RC.CNs.CI[0].T == "No Results" {
			fmt.Println("No results for ", title)
		}
	} else {
		fmt.Printf("Found [%d] results:", responseXML.RC.CNs.N)
		for _, ci := range responseXML.RC.CNs.CI {
			fmt.Printf("Found [%s] for [%s]\n", ci.Mtinfo.Mt, title)
		}
	}
}

// Main loop to process files
func main() {
	// get the working folder
	workingFolder := prompt("Working Folder", "workingFolder")

	// check if the folder exists
	fileList := getXmlSIFiles(workingFolder)

	for _, file := range fileList {
		// for each XML file in this folder
		fmt.Println("Processing: " + file)

		// Load the XML
		// Get the list of titles in nodes xpath='/ListingGroup/Listing[...]/short_event_descriptor/Title'
		titles := loadXmlFile(file)

		fmt.Println("Searching Veveo for the titles...")
		// For each title
		for _, title := range titles {
			// do search
			searchResponseXmlString := search(title)
			// for each search result
			analyseResults(title, searchResponseXmlString)
			// check the nodes xpath='/VtvRsps/RC/CNs/CI[nn]/mtinfo'
			// where <mtfn>Title</mtfn> == 'Title'
			// and <mt><![CDATA[ワンチョ -<em>伝</em>説の英雄-]]></mt> matches the original search term
			// remote <em> and </em> tags first

		}

	}

}
