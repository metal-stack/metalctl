package cmd

import "log"

func must(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
