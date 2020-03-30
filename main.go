package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
) //Notes: Tasks = a line of input from txt file

var initialTaskQ = list.New() //Initial placeholder list for input lines from txt file
//Not really important other than holds strings in a queue like fashion
//until the 'Producer' loads them all onto the 'tasks' channel queue

var tasks = make(chan Task, 10)               //Tasks channel queue
var completedTasks = make(chan Completed, 10) //Completed Tasks channel queue

type Task struct { //Struct for a task.
	textdata string //a Task has a single line of input from the txt file,
	taskid   int    //and a unique ID (just the line number)
}

type Completed struct { //Struct for a completed Task (After it's been 'consumed')
	wordcount int    //The word count from the Task's input line
	textdata  string // the input line
	taskid    int    // its unique ID
}

func createConsumers(numComsumers int) { //Creates the specified number of 'Consumers'
	var wg sync.WaitGroup               //A WaitGroup waits for a collection of goroutines to finish
	for i := 0; i < numComsumers; i++ { // This loop starts up the specified number of 'Consumers'
		wg.Add(1)        //Adds 1 to the WaitGroup variable
		go consumer(&wg) //Begins a goroutine(similar to thread) that begins to 'consume'
	}
	wg.Wait()             //Waits for all of the goroutines created above to terminate
	close(completedTasks) //Once all of the consumer goroutines have completed then nothing else
	//needs to be added to completedTasks channel queue, so close it.
}

func consumer(wg *sync.WaitGroup) { //This function actually does the 'consuming', takes a WaitGroup pointer..
	//..so that there aren't multiple WaitGroups

	for task := range tasks { //Begin looping through the 'tasks' channel queue and begin working on them
		output := Completed{wordCount(task.textdata), task.textdata, task.taskid}
		completedTasks <- output //Create a completed struct that holds all the necessary data including word count
		//Then add that Completed object to the completedTasks channel queue
	}
	wg.Done() //signals that the consumer has completed
}

func produceTask() { //This function initializes "Task" structs and ..
	//puts them onto the 'tasks' channel queue

	for i := 0; initialTaskQ.Len() > 0; i++ { //This just loops through the initialTaskQueue and builds structs
		e := initialTaskQ.Front()              //Goes in FIFO order
		str := fmt.Sprintf("%v", e.Value)      //So that the value comes back as string
		task := Task{textdata: str, taskid: i} //Create the struct
		tasks <- task                          //add task to the channel queue: 'tasks'
		initialTaskQ.Remove(e)                 //Then free up the InitialTaskQ
	}
	close(tasks) //Once all 'Tasks' have been put on the Queue, close the channel
	//AKA no more tasks to produce/consume
}

func wordCount(line string) int { //Function to count the words in a given string
	re := regexp.MustCompile(`[\S]+`)
	results := re.FindAllString(line, -1)
	return len(results) //returns count
}

func readCompleted(done chan bool) { //read out the completed tasks
	wordtotal := 0                     //variable for total word count of all 'tasks'/lines of input
	for task := range completedTasks { //loop through the completedTasks channel queue
		fmt.Printf("| Task ID: %d | word count: %d | %s | \n", task.taskid, task.wordcount, task.textdata)
		wordtotal += task.wordcount //print out the ID, the word count, and finally the text data
		//increment the wordtotal counter for final sum
	}
	fmt.Printf("| Total Word Count: %d | \n", wordtotal) //once its finished looping print out the word total
	done <- true                                         //set done to true
}

func main() {

	if len(os.Args) != 3 { //Handles the Command Line Arguments
		fmt.Println("\nUsage:", "<filename.txt>,", "<# of Consumers>") //..
		return                                                         //..
	}

	filename := os.Args[1] // file path is the 2nd argument
	// the code file name is the 1st in command line
	numberOfConsumers, err := strconv.Atoi(os.Args[2]) //Convert 3rd argument, number of Consumers
	//to an int and store in variable

	if err != nil { //Error Handling
		fmt.Println("\nmust enter number for <# of Consumers>")      //..
		fmt.Println("Usage:", "<filename.txt>,", "<# of Consumers>") //..
		os.Exit(1)                                                   //..
	}

	// This is to read in each line from input file
	fptr := flag.String("fpath", filename, "file path to read from") //..
	flag.Parse()                                                     //..

	f, err := os.Open(*fptr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}() //..

	s := bufio.NewScanner(f)    //..
	for i := 0; s.Scan(); i++ { //Keep reading in lines
		initialTaskQ.PushBack(s.Text()) //add them to the initialTaskQ
	}

	err = s.Err()   //some error handling
	if err != nil { //..
		log.Fatal(err) //..
	} //..

	go produceTask()        //create goroutine to run produceTask method
	done := make(chan bool) //create done boolean channel used to communicate
	//between goroutines safely

	go readCompleted(done)             //Create goroutine to run readCompleted method
	createConsumers(numberOfConsumers) //create the consumers
	<-done
}
