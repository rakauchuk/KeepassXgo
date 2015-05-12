package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"encoding/xml"
	"encoding/csv"
)

type Keepass struct {
	Meta struct {
	} `xml:"Meta"`
	Root struct {
		Groups    []Group    `xml:"Group"`
	} `xml:"Root"`
}

type AttrMap map[string]string
type AttrMapRaw []struct {
	Key        string
	Value      string
}

type Entry struct {
	Attributes    AttrMap    `xml:"String"`
}

type Group struct {
	Name           string
	Entries        []Entry    `xml:"Entry"`
	Groups         []Group    `xml:"Group"`
}

func (attr *AttrMap) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {

	var raw AttrMapRaw

	err := decoder.DecodeElement(&raw, &start)
	if err != nil {
		return err
	}

	if *attr == nil {
		*attr = make(AttrMap, len(raw))
	}

	for _, v := range raw {
		(*attr)[v.Key] = v.Value
	}

	return nil
}

func parse(path string, g *Group, kps *Keepass, r *[][]string) error {

	for _, e := range g.Entries {
		row := []string{ path, e.Attributes["Title"], e.Attributes["URL"], e.Attributes["UserName"], e.Attributes["Password"], e.Attributes["Notes"] }
		*r = append(*r, row)
	}

	for gr := range g.Groups {
		err := parse(path+"/"+g.Groups[gr].Name, &g.Groups[gr], kps, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {

	var kps Keepass

	var records = [][]string{{""}}

	input := "." + string(filepath.Separator) + "in"

	d, err := os.Open(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Reading files from" + input)

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".xml" {
				xmlFile, err := os.Open("in/" + file.Name())
				if err != nil {
					fmt.Println("Error opening file:", err)
					return
				}

				defer xmlFile.Close()

				csvFile, err := os.Create("out/" + file.Name() + ".csv")

				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				defer csvFile.Close()

				xmlData, _ := ioutil.ReadAll(xmlFile)

				err = xml.Unmarshal(xmlData, &kps)
				if err != nil {
					fmt.Println("Failed to process xml:", err)
				}

				fmt.Println(file.Name() + " is parsing...")

				for k := range kps.Root.Groups {

					path := kps.Root.Groups[k].Name
					err := parse(path, &kps.Root.Groups[k], &kps, &records)
					if err != nil {
						fmt.Println("Error with parser:", err)
					}
				}

				writer := csv.NewWriter(csvFile)
				for _, record := range records {
					err := writer.Write(record)
					if err != nil {
						fmt.Println("Error csv writer:", err)
						return
					}
				}
				writer.Flush()
			}
		}
	}

	fmt.Println("Done!")
}
