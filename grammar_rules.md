[Rules]
1. keyvalue pair data storage
2. root should always be an object
3. assignment is done using `=`
4. multiline string are supported using triple quotes
5. a. {"alice"}; b. {"bob", age = 50;} a is a valid usage but b is not raw values in any level are only allowed if they are in a list like ["apple","ball","cat"] or as a single value they can't be grouped with other properties if more than 1 value is at the object root then all values must have a key.
6. nesting is done using the a = {b = c = { d = "hello"} } and when writing to it using a.b.c = "hi" then the required keys are created. 
7. don't have array as a top level object like in JSON so [{},{}] is invalid for it {[{},{}]}
8. when decoding if we do like var data = dsl.decode(dataBuff); then the top level of the object is given to the variable so in the example given in rule 7 would be read using data[n].field 
9. it also supports n-D lists which can be read in conventional manner like matrix[i][j][...];
10. scope is defined using {} and values are terminted using ; 
11. each key can have just one value of type. 
12. Lists of mixed types are allowed they are lists not arrays.
13. i have a format for structured data wiz. key = (header:rows); 
example : users = (id,name,age:01,"alice",20;02,"Bob",23;);
14. whitespaces are insignificant.
15. only strings are required to be in quotes. chars can be in single quotes or double and strings have to be in double quotes.
16. empty fields are allowed and when marshalled would be given the "/0" null character.
17. the terminating semicolon is optional.
18. There are zero values for all types in tables if the headers are ineferd types from the struct then they are given else the default value is "" for all types. 
int = 0;
float = 0.0;
string = "";
bool = false;
if unsure about data type then put the '\0' as the value
final complete example of the data object

{
status = 200;
request = "POST";
error = nil;
errorCode = ;
data = {
roles = ["admin","super"];
users = (id,name,age:01,"alice",20;02,"Bob",23;);
}
}

what do you think i should name the language or data notation using these grammar rules
