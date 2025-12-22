package main

import "github.com/user/god"

import (
	"fmt"
	"log"
)

func testBareTable() {
	fmt.Println("\n=== Bare Table Test ===\n")

	// Test 1: Encode struct slice as bare table
	fmt.Println("1. Encoding []Person as bare table:")
	people := []Person{
		{Name: "John", Age: 12, Address: ""},
		{Name: "Alice", Age: 25, Address: "Boston"},
		{Name: "Bob", Age: 30, Address: "Chicago"},
	}

	encoded, err := god.MarshalBeautify(people)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encoded))

	// Test 2: Decode bare table back to []Person
	fmt.Println("\n2. Decoding bare table to []Person:")
	godData := []byte(`{(name,age,addr:"John",12,;"Alice",25,"Boston";"Bob",30,"Chicago";)}`)
	
	var decodedPeople []Person
	err = god.Unmarshal(godData, &decodedPeople)
	if err != nil {
		log.Fatal(err)
	}
	
	for i, p := range decodedPeople {
		fmt.Printf("  Person %d: %+v\n", i+1, p)
	}

	// Test 3: Compact encoding
	fmt.Println("\n3. Compact bare table:")
	compact, err := god.Marshal(people)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(compact))

	// Test 4: Round trip
	fmt.Println("\n4. Round trip test:")
	var roundTrip []Person
	err = god.Unmarshal(compact, &roundTrip)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Original: %+v\n", people)
	fmt.Printf("  Decoded:  %+v\n", roundTrip)
	
	// Verify
	match := true
	if len(people) != len(roundTrip) {
		match = false
	} else {
		for i := range people {
			if people[i] != roundTrip[i] {
				match = false
				break
			}
		}
	}
	if match {
		fmt.Println("  ✓ Round trip successful!")
	} else {
		fmt.Println("  ✗ Round trip failed!")
	}
}
