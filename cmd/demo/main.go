package main

import (
	"fmt"
	"log"
	"github.com/vinayakgupta29/god"
)

// Example struct definitions
type Person struct {
	Name    string `god:"name"`
	Age     int    `god:"age"`
	Address string `god:"addr"`
}

func main() {
	fmt.Println("=== GOD (Grounded Object Data) Encoder/Decoder Demo ===\n")

	// Example 1: Single struct encoding
	fmt.Println("1. Single Person Struct:")
	person := Person{
		Name:    "John",
		Age:     12,
		Address: "New York",
	}

	encoded, err := god.MarshalBeautify(person)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encoded:")
	fmt.Println(string(encoded))

	// Example 2: Slice of structs (table format)
	fmt.Println("\n2. Person Slice (Table Format):")
	people := []Person{
		{Name: "John", Age: 12, Address: ""},
		{Name: "Alice", Age: 25, Address: "Boston"},
		{Name: "Bob", Age: 30, Address: "Chicago"},
	}

	encodedSlice, err := god.MarshalBeautify(people)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encoded:")
	fmt.Println(string(encodedSlice))

	// Example 3: Decoding single struct
	fmt.Println("\n3. Decoding Single Person:")
	godData := []byte(`{name="Jane";age=28;addr="Seattle"}`)
	var decodedPerson Person
	err = god.Unmarshal(godData, &decodedPerson)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decoded: %+v\n", decodedPerson)

	// Example 4: Decoding table format
	fmt.Println("\n4. Decoding Person Slice (Table Format):")
	tableGOD := []byte(`{(name,age,addr:"Alice",30,"NYC";"Bob",25,"LA";)}`)
	var decodedPeople []Person
	err = god.Unmarshal(tableGOD, &decodedPeople)
	if err != nil {
		log.Fatal(err)
	}
	for i, p := range decodedPeople {
		fmt.Printf("  Person %d: %+v\n", i+1, p)
	}

	// Example 5: Compact encoding
	fmt.Println("\n5. Compact Encoding:")
	compact, err := god.Marshal(person)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Compact:", string(compact))

	// Example 6: Map encoding
	fmt.Println("\n6. Map Encoding:")
	data := map[string]interface{}{
		"status":  200,
		"message": "Success",
		"data": map[string]interface{}{
			"count": 3,
			"items": []interface{}{"apple", "banana", "cherry"},
		},
	}
	mapEncoded, err := god.MarshalBeautify(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(mapEncoded))

	// Example 7: Grounded Nulls (\0)
	fmt.Println("\n7. Grounded Nulls (\\0):")
	nullData := []byte(`{errorCode=\0;errorMessage=}`)
	var result map[string]interface{}
	err = god.Unmarshal(nullData, &result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decoded map with \\0: {errorCode=%q;errorMessage=%q}\n", result["errorCode"], result["errorMessage"])

	testRule5Examples()
	testBareTable()
	example2()

	fmt.Println("\n=== Demo Complete ===")
}

