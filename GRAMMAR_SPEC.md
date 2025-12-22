# GOD (Grounded Object Data) Formal Grammar Specification

**Version:** 1.0.0  
**Date:** December 2025  
**Status:** Final Draft

## 1. Introduction

GOD (Grounded Object Data) is a lightweight, human-readable data serialization format designed as a more compact and safer alternative to JSON. It provides native support for tabular data and enforces "grounded" semantics where every data point corresponds to a concrete value (zero values) rather than undefined or null states.

### 1.1 Design Goals

- **Grounding**: All data points are grounded in concrete types. There is no `null` value; omitted fields return the type's zero value.
- **Compactness**: Significantly reduce data size compared to JSON, especially for repetitive structured data.
- **Readability**: Maintain a syntax that is easy for humans to read and write.
- **Tabular Power**: First-class support for columnar/tabular data.
- **Space Efficiency**: Optimized for storage and transmission.

### 1.2 Key Differences from JSON

| Feature | GOD | JSON |
|---------|-----|------|
| Null values | ❌ Omitted = Zero values | ✅ Explicit null |
| Tabular data | ✅ Native `(header:rows)` | ❌ Array of objects |
| Keys | ✅ Optional for single values | ✅ Always required |
| Root requirement | Must be object `{}` | Any type |
| Safety | ✅ Safeguard against null | ⚠️ Allows null values |

## 2. Lexical Structure

### 2.1 Character Set

GOD uses UTF-8 encoding.

### 2.2 Whitespace

Whitespace characters (space, tab, newline, carriage return) are insignificant except within string literals.

### 2.3 Tokens

```
token ::= '{' | '}' | '[' | ']' | '(' | ')' | '=' | ';' | ',' | ':' | string | number | boolean | identifier
```

## 3. Grammar Rules

### 3.1 Root Structure

**Rule 1**: The root element MUST be an object enclosed in braces `{}`.

**Rule 2**: If the object root contains more than one value, all values MUST have a key.

**Invalid:**
```
{"John"}         ✅ Valid (single value)
{"John", 25}     ❌ Invalid (multiple naked values)
{name="John", 25} ❌ Invalid (mixed keyed and naked)
```

**Valid:**
```
{name="John"; age=25} ✅ Valid (all keyed)
```

### 3.2 Objects

An object contains content enclosed in braces.

```ebnf
object ::= '{' content '}'
content ::= raw-value | key-value-pairs | empty
raw-value ::= value
key-value-pairs ::= key-value-pair (term key-value-pair)* term?
term ::= ';' | whitespace+
```

### 3.3 Values

```ebnf
value ::= string | number | boolean | object | array | table | empty
```

### 3.4 Key-Value Assignment

Assignment uses the `=` operator.

```ebnf
key-value-pair ::= identifier '=' value
identifier ::= [a-zA-Z_][a-zA-Z0-9_]*
```

### 3.5 Strings and Characters

- **Strings**: Must be in double quotes `"`.
- **Characters**: Can be in single quotes `'` or double quotes `"`.
- **Multiline**: Supported using triple quotes `"""`.

```ebnf
string ::= '"' char* '"' | '"""' multiline-char* '"""'
```

### 3.6 Arrays (Lists)

Arrays contain comma-separated values in brackets `[]`. Mixed types are allowed.

```ebnf
array ::= '[' (value (',' value)*)? ']'
```

### 3.7 Tables (Tabular Structured Data)

**Rule 13**: Tables provide an efficient way to store lists of similar objects.

```ebnf
table ::= '(' header ':' rows ')'
header ::= identifier (',' identifier)*
rows ::= row (';' row)* ';'?
row ::= cell (',' cell)*
cell ::= value
```

**Example:**
```
users = (id,name,age:01,"alice",20;02,"Bob",23;);
```

## 4. Grounding and Zero Values

**Rule 18**: The core philosophy of GOD is that every field is grounded. When data is missing or empty, it is automatically assigned the type's zero value.

| Type | Zero Value | Marshalling |
|------|------------|-------------|
| **Integer** | `0` | `0` |
| **Float** | `0.0` | `0.0` |
| **String** | `""` | `""` |
| **Boolean** | `false` | `false` |
| **Unknown** | `\0` | Null character |

If a field is empty (e.g., `errorCode = ;`), it is marshalled with the `\0` character.

## 5. Usage Example

```
{
  status = 200;
  request = "POST";
  error = "";
  errorCode = ;
  data = {
    roles = ["admin", "super"];
    users = (id,name,age:01,"alice",20;02,"Bob",23;);
  }
}
```

## 6. License

GOD is free for personal, creative, and educational use. Commercial use requires prior permission from the developer. See the [LICENSE](LICENSE) file for full terms.
