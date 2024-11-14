package main

import (
	"log"
	"os"

	"github.com/ejuju/nudoc/pkg/nudoc"
)

func main() {
	f, err := os.Open("examples/demo/intro.nudoc")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	doc, err := nudoc.Parse(f)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Title: %s", doc.Title)
	for _, n := range doc.Nodes {
		log.Printf("%T %#v", n, n)
	}

	err = nudoc.WriteHTML(os.Stdout, doc)
	if err != nil {
		log.Fatal(err)
	}
}
