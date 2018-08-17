package regex

import( 
	"fmt"
	"sort"
	"os"
	"bufio"
	"log"
)


//The struct type of the transition lookup
type transition struct {
	id int 
	trans string
}


// Remove from s at index i
func remove(s []int, i int) []int {
    return append(s[:i],s[i+1:]...)
}

//Returns true if list s contains i
func contains(s []int, i int) bool {
	for _, v := range s {
		if v == i {
			return true
		}
	}
	return false
}

//Split the regular expression into an array of elements
//Inputs:
//	regex: the regex to split
//Output:
//	A string array of the individual regex elements
func splitString(regex string) []string{
	//Empty string slice
	splits := make([]string, 0)
	//The current element (in this grammar it will either be 1 or 2 elements)
	var curr string
	//Did we add a character last?
	prev := false
	//Loop over the expression
	for i, c := range regex {
		//Are we at a modifier?
		if c == '*' || c == '+' {
			//If we didn't previously have a character we have two modifiers in a row -- error
			if !prev {
				log.Panic("Regular expression parsing failed")
			}
			//Add the modifer to the string
			curr += string(c)
			//Append the token to the list
			splits = append(splits,curr)
			//This wasn't a character
			prev = false
		} else {
			//Last token was a character, and we have no modifers, lets add it.
			if prev {
				splits = append(splits,curr)
			}
			//This was a character, set the variable to true and loop over to check for modifiers
			curr = string(c)
			prev = true
		}
		//We're a the last character and it's not a modifier.
		if i == len(regex) - 1 && prev {
			splits = append(splits,curr)
		}
		
	}
	return splits
}
// Lookup fsm[transition] returns the list of the next nodes IDs -- used to create next transition
// transition.trans == "" means to not consume the character
var fsm map[transition][]int
//To parse the grammar we're using a Finite State Machine (FSM)
//The FSM will be created from the tokenized list of the grammar created above
//The FSM will be stored in a map of transitions to a list of state numbers
//Inputs:
//	pieces: The tokenized grammar from above
//Returns
//	An int declaring the final statte number
func createFsm(pieces []string) int{
	//Initialize the map we declared above
	fsm = make(map[transition][]int)
	//This is the final state number
	var last int
	//Loop over the tokens
	for i, next := range pieces {
		//Create our transition
		n := transition {
			id: i,
			trans: string(next[0]),
		}
		//Initialize the FSM
		fsm[n] = make([]int,0)
		//If we have a simple token, we just transition to the next state
		if len(next) == 1 {
			fsm[n] = []int{i + 1}
		} else {
			//Grab the modifier
			mod := next[1]
			//If we have the + modifier
			if mod == '+' {
				//We can transition to self as well as to the next state
				fsm[n] = []int{i, i+1}
			//If we have the * modifier
			} else if mod == '*' {
				//Create a new transition for when we don't consume the token
				nSt := transition {
					id: i,
					trans: "", //Empty string indicates nonconsuming transition
				}
				//We either move to the next node without consuming or move to the current node and consume
				fsm[nSt] = []int{i + 1}
				fsm[n] = []int{i}
			}
		}
		//Return the number of the final state
		last = i
	}
	return last + 1
}

//Parse the input with the regular expression.
//This was originally recursive grammar parsing, but that got crazy slow with larger inputs.
//That's when the idea for FSMs came in.
//This has been optimized on top of that to allow the FSM to be in multiple states, brancing when allowed.
//Inputs
//	regex: The regular expression to match
//	match: The string to match the regex to
//Outputs:
//	Two slices
//	{0}: The actual text of each match
//	{1}: The start index of the match
//	{0} and {1} have the same length
func parse(regex string, match string) ([]string,[]int){
	//Create our FSM
	last := createFsm(splitString(regex))
	//Create our return variables
	matches := make([]string, 0)
	index := make([]int,0)
	//Loop over the input
	for start := 0; start < len(match); start++ {
		//Our state buffer (we allow mutiple states)
		currentStates := make([]int, 1)
		//The end of a successful match
		end := start
		//Finished a match
		var done bool
		//Loop over the rest of the string parsing
		for let := start; let < len(match) && !done; {
			//We aren't done when we get to a new character
			done = false
			//Loop over our concurrent states
			for state := 0; state < len(currentStates) && !done;  {
				//If we've hit the end of the text, let the list loop one more time
				if let >= len(match){let--;  done=true}
				//We are a lazy parser, we want to parse the furthest along first
				sort.Sort(sort.Reverse(sort.IntSlice(currentStates)))
				//Simple transition
				next := transition {
					id: currentStates[state],
					trans: string(match[let]),
				}
				//Nonconsuming transition
				nextNil := transition {
					id: currentStates[state],
					trans: "",
				}
				//Wildcard transition
				nextAny := transition {
					id: currentStates[state],
					trans: ".",
				}
				
				//Simple and wildcard next states
				add := fsm[next]
				addAny := fsm[nextAny]
				
				//If we have some new states
				if add != nil {
					//Add them
					currentStates = append(currentStates,add...)
					//And consume
					let++
				}
				if addAny != nil {
					//Add
					currentStates = append(currentStates, addAny...)
					//Consume
					if add == nil {let++}
					
				}
				addNil := fsm[nextNil]
				if addNil != nil {
					//Add states and don't consume
					currentStates = append(currentStates, addNil...)
					
				}
				
				//We've parsed this state
				currentStates = remove(currentStates, state)
				//Nothing left to parse
				if len(currentStates) == 0 {
					done = true
				}
		
				//We've seen the final node, we've a successful match
				if contains(currentStates,last) {
					
					done = true
					end = let
					matches = append(matches,match[start:end])
					index = append(index,start)
					start = end-1
				}
			
			}
		}
	}
	//Return
	return matches,index
}


//Node struct to be used in v2
type node struct {
	//The character to transition on
	trans rune
	//The pointers to the next nodes. If n2 isn't nill we branch
	//If n1 and n2 are nil, we're done
	n1 *node
	n2 *node
}

//Takes a tokenized list of the regex grammar and returns a pointer to the starting node of the FSM
func createFsmV2(pieces []string) *node {
	start := &node{
		trans: ''
		n1: nil
		n2: nil
	}
	for i, piece := range pieces {
		//If we have no modifier
		if len(piece) == 1 {
			
		}
	}
	return start
}

//Takes a regex string and a search strings and returns all of the matches of regex in match
func parseV2(regex string, match string) ([]string, []int) {
	_ := createFsm(splitString(regex))
	return make([]string,0),make([]int,0)
}

//The Match function is visiable outside the package and is the way to match a regex to a file
//This function could be spead up slightly more by putting each lines parsing in a goroutiene
//To make the goroutine work, one would have to use a worker pool and a routiene dispacher.
//Inputs:
//	file: The file to parse
//	regex: The regular expression to match
//Returns:
//	A newline separated string of the matches
func Match(file *os.File,regex string) string{
	//Read a buffered file
	scan := bufio.NewScanner(file)
	//Return variable
	ret :=""
	//Current line number
	lineNo := 0
	//Loop over the file
	for scan.Scan(){
		lineNo++
		//Read in the line and parse the expression then add the results to our return variable.
		line := scan.Text()
		p,in:=parseV2(regex,line)
		for i,v := range p {
			ret = ret + fmt.Sprintf("%s matches (line: %d, index: %d)\n",v,lineNo,in[i])
		} 
	}
	return ret
	
}