package main

import (
	"log"
	"os"

	"github.com/ejuju/nudoc/pkg/nudoc"
)

func main() {
	f, err := os.Open("doc/design.nudoc")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	doc, err := nudoc.ParseDocument(nudoc.NewReader(f))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Name: %s", doc.Header.Name)
	for _, n := range doc.Body.Nodes {
		log.Printf("%T %#v", n, n)
	}

	err = nudoc.WriteHTML(os.Stdout, doc)
	if err != nil {
		log.Fatal(err)
	}
}
