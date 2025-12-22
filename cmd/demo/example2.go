package main

import "github.com/vinayakgupta29/god"

import (
	"fmt"
	"log"
)

func example2() {
	fmt.Println("\n=== Additional Examples ===\n")

	// Example 1: Struct with slice of structs (nested table)
	fmt.Println("1. Company with Employees (Nested Structure):")
	company := map[string]interface{}{
		"name":    "TechCorp",
		"founded": 2020,
		"employees": []Person{
			{Name: "Alice", Age: 30, Address: "NYC"},
			{Name: "Bob", Age: 25, Address: "LA"},
			{Name: "Charlie", Age: 35, Address: ""},
		},
	}

	encoded, err := god.MarshalBeautify(company)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encoded))

	// Example 2: Just a slice (gets wrapped with "data" key)
	fmt.Println("\n2. Plain Slice (Auto-wrapped in object):")
	numbers := []interface{}{1, 2, 3, 4, 5}
	encodedNums, err := god.MarshalBeautify(numbers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encodedNums))

	// Example 3: Compact encoding of slice
	fmt.Println("\n3. Compact Slice Encoding:")
	compactNums, err := god.Marshal(numbers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(compactNums))

	// Example 4: Decoding the company structure
	// TODO: Fix table decoding issue
	/*
	fmt.Println("\n4. Decoding Company with Employee Table:")
	godCompany := []byte(`{name="MegaCorp";founded=2015;employees=(name,age,addr:"John",28,"Boston";"Jane",32,"Seattle";)}`)

	var result map[string]interface{}
	err = god.Unmarshal(godCompany, &result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decoded: %+v\n", result)
	*/
}
