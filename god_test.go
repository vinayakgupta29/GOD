package god

import (
	"fmt"
	"strings"
	"testing"
)

type Person struct {
	Name    string `god:"name"`
	Age     int    `god:"age"`
	Address string `god:"addr"`
}

// Additional test struct
type Company struct {
	Name      string   `god:"name"`
	Founded   int      `god:"founded"`
	Employees []Person `god:"employees"`
}

type Response struct {
	Status    int                    `god:"status"`
	Request   string                 `god:"request"`
	Error     string                 `god:"error"`
	ErrorCode string                 `god:"errorCode"`
	Data      map[string]interface{} `god:"data"`
}


func TestSinglePersonEncode(t *testing.T) {
	person := Person{
		Name:    "John",
		Age:     12,
		Address: "New York",
	}

	// Compact encoding
	compact, err := Marshal(person)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	fmt.Println("=== Single Person (Compact) ===")
	fmt.Println(string(compact))
	fmt.Println()

	// Beautified encoding
	pretty, err := MarshalBeautify(person)
	if err != nil {
		t.Fatalf("MarshalBeautify error: %v", err)
	}
	fmt.Println("=== Single Person (Beautified) ===")
	fmt.Println(string(pretty))
	fmt.Println()
}

func TestPersonSliceEncode(t *testing.T) {
	people := []Person{
		{Name: "John", Age: 12, Address: ""},
		{Name: "Alice", Age: 25, Address: "Boston"},
		{Name: "Bob", Age: 30, Address: "Chicago"},
	}

	// Compact encoding (should use table format)
	compact, err := Marshal(people)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	fmt.Println("=== Person Slice (Compact - Table Format) ===")
	fmt.Println(string(compact))
	fmt.Println()

	// Beautified encoding
	pretty, err := MarshalBeautify(people)
	if err != nil {
		t.Fatalf("MarshalBeautify error: %v", err)
	}
	fmt.Println("=== Person Slice (Beautified - Table Format) ===")
	fmt.Println(string(pretty))
	fmt.Println()
}

func TestPersonDecode(t *testing.T) {
	dslData := []byte(`{name="John";age=12;addr="New York"}`)

	var person Person
	err := Unmarshal(dslData, &person)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	fmt.Println("=== Decoded Person ===")
	fmt.Printf("Name: %s, Age: %d, Address: %s\n", person.Name, person.Age, person.Address)
	fmt.Println()

	if person.Name != "John" || person.Age != 12 || person.Address != "New York" {
		t.Errorf("Decoded values don't match expected")
	}
}

func TestPersonSliceDecode(t *testing.T) {
	dslData := []byte(`{(name,age,addr:"John",12,"";)}`)

	var people []Person
	err := Unmarshal(dslData, &people)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	fmt.Println("=== Decoded Person Slice ===")
	for i, p := range people {
		fmt.Printf("Person %d: Name=%s, Age=%d, Address=%s\n", i+1, p.Name, p.Age, p.Address)
	}
	fmt.Println()

	if len(people) != 1 {
		t.Errorf("Expected 1 person, got %d", len(people))
	}
	if people[0].Name != "John" || people[0].Age != 12 {
		t.Errorf("Decoded values don't match expected")
	}
}

func TestComplexStructEncode(t *testing.T) {
	company := Company{
		Name:    "TechCorp",
		Founded: 2020,
		Employees: []Person{
			{Name: "Alice", Age: 30, Address: "NYC"},
			{Name: "Bob", Age: 25, Address: "LA"},
		},
	}

	// Compact encoding
	compact, err := Marshal(company)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	fmt.Println("=== Company (Compact) ===")
	fmt.Println(string(compact))
	fmt.Println()

	// Beautified encoding
	pretty, err := MarshalBeautify(company)
	if err != nil {
		t.Fatalf("MarshalBeautify error: %v", err)
	}
	fmt.Println("=== Company (Beautified) ===")
	fmt.Println(string(pretty))
	fmt.Println()
}

func TestMapEncode(t *testing.T) {
	data := map[string]interface{}{
		"status":  200,
		"message": "OK",
		"data": map[string]interface{}{
			"users": []interface{}{"alice", "bob"},
			"count": 2,
		},
	}

	// Compact encoding
	compact, err := Marshal(data)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	fmt.Println("=== Map (Compact) ===")
	fmt.Println(string(compact))
	fmt.Println()

	// Beautified encoding
	pretty, err := MarshalBeautify(data)
	if err != nil {
		t.Fatalf("MarshalBeautify error: %v", err)
	}
	fmt.Println("=== Map (Beautified) ===")
	fmt.Println(string(pretty))
	fmt.Println()
}

func TestRoundTrip(t *testing.T) {
	original := []Person{
		{Name: "Alice", Age: 30, Address: "NYC"},
		{Name: "Bob", Age: 25, Address: ""},
	}

	// Encode
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	fmt.Println("=== Round Trip Test ===")
	fmt.Println("Encoded:", string(encoded))

	// Decode
	var decoded []Person
	err = Unmarshal(encoded, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	fmt.Printf("Decoded: %+v\n", decoded)
	fmt.Println()

	// Verify
	if len(decoded) != len(original) {
		t.Errorf("Length mismatch: expected %d, got %d", len(original), len(decoded))
	}
	for i := range original {
		if decoded[i].Name != original[i].Name || decoded[i].Age != original[i].Age || decoded[i].Address != original[i].Address {
			t.Errorf("Person %d mismatch: expected %+v, got %+v", i, original[i], decoded[i])
		}
	}
}

func TestOptionalSemicolon(t *testing.T) {
	// Test with semicolons
	dslWithSemi := []byte(`{name="John";age=12}`)
	var p1 Person
	err := Unmarshal(dslWithSemi, &p1)
	if err != nil {
		t.Fatalf("Unmarshal with semicolon error: %v", err)
	}

	// Test without semicolons (rule 17)
	dslWithoutSemi := []byte(`{name="Jane" age=15}`)
	var p2 Person
	err = Unmarshal(dslWithoutSemi, &p2)
	if err != nil {
		t.Fatalf("Unmarshal without semicolon error: %v", err)
	}

	fmt.Println("=== Optional Semicolon Test ===")
	fmt.Printf("With semicolon: %+v\n", p1)
	fmt.Printf("Without semicolon: %+v\n", p2)
	fmt.Println()
}

func TestEmptyFields(t *testing.T) {
	person := Person{
		Name:    "John",
		Age:     0,
		Address: "",
	}

	encoded, err := MarshalBeautify(person)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	fmt.Println("=== Empty Fields Test ===")
	fmt.Println(string(encoded))
	fmt.Println()
}

func TestBeautifyRootIndention(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}
	encoded, _ := MarshalBeautify(data)
	s := string(encoded)
	if !strings.Contains(s, "\n  key=") {
		t.Errorf("Expected indentation level 1 for root items, got:\n%s", s)
	}
}

func TestTableBeautify(t *testing.T) {
	people := []Person{
		{Name: "John", Age: 30, Address: "NYC"},
		{Name: "Alice", Age: 25, Address: "Boston"},
	}
	encoded, _ := MarshalBeautify(people)
	s := string(encoded)
	fmt.Println("=== Table Beautify Test ===")
	fmt.Println(s)
	
	expectedPart := "(name,age,addr:\n  \"John\",30,\"NYC\";\n  \"Alice\",25,\"Boston\";\n)"
	if !strings.Contains(s, expectedPart) {
		t.Errorf("Table beautify formatting incorrect. Expected part:\n%s\nGot:\n%s", expectedPart, s)
	}
}
