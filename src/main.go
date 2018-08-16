
package main

import (
	"regex"
	"os"
	"log"
)
//Run the main method
func main(){
	//If we don't have the right command
	if len(os.Args) < 2{
		//Fail
		log.Fatal("Not enough arguments. Please enter a file followed by a regular expression")
	}
	//Fine name
	fName := os.Args[1]
	//Regular expression
	rExp := os.Args[2]
	//Try to open the file
	f, err := os.Open(fName)
	//If we cant,
    if err != nil {
		//fail
        log.Fatal(err)
	}
	//Defer the file closing and error handling if it doesn't close
    defer func() {
        if err = f.Close(); err != nil {
        log.Fatal(err)
    }
	}()
	//Search for the expression in the lines of the file
	log.Printf("Beginning search for %s in file %s\n",rExp,fName)
	log.Println(regex.Match(f,rExp))
}
