# GOD - Grounded Object Data

GOD (Grounded Object Data) is a lightweight, human-readable data serialization format designed as a compact and safe alternative to JSON. 

The core philosophy of GOD is that data should be **grounded**—there are no ambiguities or null/undefined states. If a data point exists, it has a value; if it's missing, it's grounded in its type's zero value.

## Why GOD?

- 🕊️ **Grounded**: Prevents null pointer errors by using zero values (`0`, `""`, `false`).
- 📦 **Compact**: Up to 50% smaller than JSON for tabular records.
- 📊 **Table Support**: Native syntax for list of objects, saving massive space on keys.
- 🛠️ **Reflection-Driven**: Easy integration with Go structs using `god` tags.
- 📖 **Human Readable**: Simple assignment and nesting that anyone can follow.

## Installation

```bash
go get github.com/vinayakgupta29/god
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/vinayakgupta29/god"
)

type User struct {
    ID   int    `god:"id"`
    Name string `god:"name"`
}

func main() {
    // Encoding a list of users
    users := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }
    
    encoded, _ := god.MarshalBeautify(users)
    fmt.Println(string(encoded))
    // Output:
    // {
    //   (id,name:1,"Alice";2,"Bob";)
    // }
}
```

## Comparisons

### JSON
```json
{
  "users": [
    {"id": 1, "name": "Alice", "age": 20},
    {"id": 2, "name": "Bob", "age": 23}
  ]
}
```

### GOD (Grounded Object Data)
```
{
  users = (id,name,age:1,"Alice",20;2,"Bob",23;)
}
```

## License

GOD is licensed under a custom license that is **free for personal, creative, and educational use**. 

⚠️ **Commercial Use**: Use for any commercial purposes requires explicit prior permission from the developer. Please see the [LICENSE](LICENSE) file for details.

## Documentation

For full details on the specification, see [GRAMMAR_SPEC.md](GRAMMAR_SPEC.md).
