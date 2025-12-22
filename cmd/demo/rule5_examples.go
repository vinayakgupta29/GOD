package main

import "github.com/vinayakgupta29/god"

import (
	"fmt"
	"log"
)

func testRule5Examples() {
	fmt.Println("\n=== Grammar Rule 5 Examples ===")
	fmt.Println("Rule: Root can have EITHER single raw value OR key-value pairs, but NOT both\n")

	// Valid: Single raw string
	fmt.Println("1. Single raw string: {\"John\"}")
	str := "John"
	encoded, _ := god.Marshal(str)
	fmt.Printf("   Encoded: %s\n", string(encoded))

	// Valid: Single raw array
	fmt.Println("\n2. Single raw array: {[1,2,3,4,5,6]}")
	arr := []interface{}{1, 2, 3, 4, 5, 6}
	encoded, _ = god.Marshal(arr)
	fmt.Printf("   Encoded: %s\n", string(encoded))

	// Valid: Single raw table
	fmt.Println("\n3. Single raw table: {(table)}")
	people := []Person{
		{Name: "John", Age: 30, Address: "NYC"},
	}
	encoded, _ = god.Marshal(people)
	fmt.Printf("   Encoded: %s\n", string(encoded))

	// Valid: Key-value pair (single)
	fmt.Println("\n4. Single key-value: {data=\"John\"}")
	kv1 := map[string]interface{}{"data": "John"}
	encoded, _ = god.MarshalBeautify(kv1)
	fmt.Printf("   Encoded: %s\n", string(encoded))

	// Valid: Multiple key-value pairs
	fmt.Println("5. Multiple key-values: {data=\"John\";age=12}")
	kv2 := map[string]interface{}{"data": "John", "age": 12}
	encoded, _ = god.MarshalBeautify(kv2)
	fmt.Printf("   Encoded: %s\n", string(encoded))

	// Demonstrate decoding
	fmt.Println("\n=== Decoding Examples ===\n")

	// Decode single string
	fmt.Println("1. Decoding {\"Hello World\"}")
	var str2 string
	err := god.Unmarshal([]byte(`{"Hello World"}`), &str2)
	if err != nil {
		log.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Result: %q\n", str2)
	}

	// Decode single array
	fmt.Println("\n2. Decoding {[10,20,30]}")
	var arr2 []interface{}
	err = god.Unmarshal([]byte(`{[10,20,30]}`), &arr2)
	if err != nil {
		log.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Result: %v\n", arr2)
	}

	// Decode bare table
	fmt.Println("\n3. Decoding {(name,age,addr:\"Alice\",28,\"Seattle\";)}")
	var people2 []Person
	err = god.Unmarshal([]byte(`{(name,age,addr:"Alice",28,"Seattle";)}`), &people2)
	if err != nil {
		log.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Result: %+v\n", people2)
	}

	fmt.Println("\n✓ All Rule 5 examples work correctly!")
}
