package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var initialtaskQ = list.New()

var tasks = make(chan Task, 10)
var completedTasks = make(chan Completed, 10)

type Task struct {
	textdata string
	taskid   int
}

type Completed struct {
	wordcount int
	textdata  string
	taskid    int
}

func consumer(wg *sync.WaitGroup) {
	for task := range tasks {
		output := Completed{wordCount(task.textdata), task.textdata, task.taskid}
		completedTasks <- output
	}
	wg.Done()
}

func createConsumers(numComsumers int) {
	var wg sync.WaitGroup
	for i := 0; i < numComsumers; i++ {
		wg.Add(1)
		go consumer(&wg)
	}
	wg.Wait()
	close(completedTasks)
}

func produceTask() {
	for i := 0; initialtaskQ.Len() > 0; i++ {
		e := initialtaskQ.Front() // First element
		str := fmt.Sprintf("%v", e.Value)
		task := Task{textdata: str, taskid: i}
		tasks <- task
		initialtaskQ.Remove(e) // Dequeue
	}
	close(tasks)
}

func wordCount(line string) int {
	strArray := strings.Split(line, " ")
	return len(strArray)
}

func readCompleted(done chan bool) {
	for task := range completedTasks {
		fmt.Printf("Task id: %d, input line: %s , word count: %d\n", task.taskid, task.textdata, task.wordcount)
	}
	done <- true
}

func main() {
	startTime := time.Now()

	fptr := flag.String("fpath", "textdata.txt", "file path to read from")
	flag.Parse()

	f, err := os.Open(*fptr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	s := bufio.NewScanner(f)
	count := 0
	for i := 0; s.Scan(); i++ {
		initialtaskQ.PushBack(s.Text())
		count++
	}

	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}

	go produceTask()
	done := make(chan bool)
	go readCompleted(done)
	createConsumers(10)
	<-done
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Println("total time taken ", diff.Seconds(), "seconds")
}
