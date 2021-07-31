package main

import (
	"bufio"
	"fmt"
	"github.com/jonnyGit81/chat/domain/thesaurus"
	"log"
	"os"
)

func main() {
	//apiKey := os.Getenv("BHT_APIKEY")
	apiKey := "e2156942a266d0349bb26f8ae7771462"
	thesaurus := &thesaurus.BigHuge{APIKey: apiKey}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		word := s.Text()
		syns, err := thesaurus.Synonyms(word)
		if err != nil {
			log.Fatalln("Failed when looking for synonyms for  "+word, err)
		}
		if len(syns) == 0 {
			log.Fatalln("Couldn't find any synonyms for " + word)
		}
		for _, syn := range syns {
			fmt.Println(syn)
		}
	}
}
