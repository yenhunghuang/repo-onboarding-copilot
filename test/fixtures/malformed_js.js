// Malformed JavaScript file for testing error handling

// Missing closing brace
function testFunction() {
    console.log("This function is missing a closing brace");
    if (true) {
        return "incomplete";
    // Missing closing brace for function

// Incomplete class declaration
class IncompleteClass {
    constructor() {
        this.value = "test"
    
    // Missing method implementation
    incompleteMethod() {
        return
    }
    // Missing closing brace for class

// Invalid syntax
const invalidAssignment = 
const anotherVar = "this should cause parsing issues"

// Unclosed string literal
const unclosedString = "This string is never closed

// Invalid object literal
const brokenObject = {
    key1: "value1",
    key2: "value2"
    key3: // Missing value
};

// Malformed arrow function
const brokenArrow = (param1, param2 => {
    return param1 + param2;
};