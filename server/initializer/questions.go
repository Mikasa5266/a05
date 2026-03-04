package initializer

import (
	"fmt"
	"log"

	"your-project/model"

	"gorm.io/gorm"
)

// InitSampleQuestions initializes sample questions for various positions
func InitSampleQuestions(db *gorm.DB) error {
	sampleQuestions := []model.Question{
		// Python questions
		{
			Title:          "Python Basics: Variables and Data Types",
			Content:        "Please explain mutable and immutable data types in Python with examples.",
			Position:       "Python",
			Difficulty:     "Junior",
			Category:       "Basics",
			Tags:           "data types,basics",
			ExpectedAnswer: "Immutable data types in Python include: int, float, str, tuple, etc. Once created, they cannot be modified. Mutable data types include: list, dict, set, etc. They can be modified in place.",
		},
		{
			Title:          "Python List Operations",
			Content:        "Please explain slicing operations in Python lists and how to reverse a list.",
			Position:       "Python",
			Difficulty:     "Junior",
			Category:       "Data Structures",
			Tags:           "lists,slicing,reverse",
			ExpectedAnswer: "List slicing syntax is list[start:stop:step]. To reverse a list, you can use list[::-1] or list.reverse() method.",
		},
		{
			Title:          "Python Function Definitions",
			Content:        "Please explain the purpose and difference between *args and **kwargs parameters in Python functions.",
			Position:       "Python",
			Difficulty:     "Junior",
			Category:       "Functions",
			Tags:           "functions,parameters",
			ExpectedAnswer: "*args is used to receive any number of positional arguments, **kwargs is used to receive any number of keyword arguments.",
		},
		{
			Title:          "Python Object-Oriented Programming",
			Content:        "Please explain the concepts of inheritance and polymorphism in Python.",
			Position:       "Python",
			Difficulty:     "Junior",
			Category:       "OOP",
			Tags:           "inheritance,polymorphism,OOP",
			ExpectedAnswer: "Inheritance allows subclasses to inherit attributes and methods from parent classes. Polymorphism allows objects of different classes to respond differently to the same message.",
		},
		{
			Title:          "Python Exception Handling",
			Content:        "Please explain the execution order of try-except-finally statements in Python.",
			Position:       "Python",
			Difficulty:     "Junior",
			Category:       "Exception Handling",
			Tags:           "exceptions,try-except",
			ExpectedAnswer: "First, the try block is executed. If there is an exception, the except block is executed. Finally, the finally block is executed regardless of whether there was an exception.",
		},
		{
			Title:          "Python Intermediate: Decorators",
			Content:        "Please explain the purpose of Python decorators and write a simple decorator example.",
			Position:       "Python",
			Difficulty:     "Intermediate",
			Category:       "Advanced Features",
			Tags:           "decorators,advanced features",
			ExpectedAnswer: "Decorators are used to modify or enhance the behavior of functions. Example: @decorator def func(): pass",
		},
		{
			Title:          "Python Generators",
			Content:        "Please explain the difference between Python generators and regular functions, and the role of the yield keyword.",
			Position:       "Python",
			Difficulty:     "Intermediate",
			Category:       "Advanced Features",
			Tags:           "generators,yield",
			ExpectedAnswer: "Generators use yield to return data and can pause and resume execution, saving memory.",
		},
		{
			Title:          "Python Advanced: Concurrent Programming",
			Content:        "Please explain the difference between multithreading and multiprocessing in Python and their applicable scenarios.",
			Position:       "Python",
			Difficulty:     "Senior",
			Category:       "Concurrency",
			Tags:           "multithreading,multiprocessing,concurrency",
			ExpectedAnswer: "Multithreading is suitable for I/O-bound tasks, multiprocessing is suitable for CPU-bound tasks.",
		},

		// JavaScript questions
		{
			Title:          "JavaScript Basics: Variables and Data Types",
			Content:        "Please explain the differences between var, let, and const in JavaScript.",
			Position:       "Frontend",
			Difficulty:     "Junior",
			Category:       "Basics",
			Tags:           "variables,data types,basics",
			ExpectedAnswer: "var has function scope, let and const have block scope. const cannot be reassigned, while var and let can.",
		},
		{
			Title:          "JavaScript Functions",
			Content:        "Please explain arrow functions in JavaScript and their differences from regular functions.",
			Position:       "Frontend",
			Difficulty:     "Junior",
			Category:       "Functions",
			Tags:           "functions,arrow functions",
			ExpectedAnswer: "Arrow functions have a shorter syntax and do not bind their own this, arguments, super, or new.target.",
		},
		{
			Title:          "JavaScript DOM Manipulation",
			Content:        "Please explain how to select elements from the DOM and modify their properties.",
			Position:       "Frontend",
			Difficulty:     "Junior",
			Category:       "DOM",
			Tags:           "DOM,manipulation",
			ExpectedAnswer: "Use methods like document.getElementById(), document.querySelector(), document.querySelectorAll() to select elements, then modify their properties or content.",
		},

		// Java questions
		{
			Title:          "Java Basics: Variables and Data Types",
			Content:        "Please explain the primitive data types in Java and their ranges.",
			Position:       "Java",
			Difficulty:     "Junior",
			Category:       "Basics",
			Tags:           "variables,data types,basics",
			ExpectedAnswer: "Java has 8 primitive data types: byte, short, int, long, float, double, char, and boolean. Each has a specific range and size.",
		},
		{
			Title:          "Java OOP Concepts",
			Content:        "Please explain encapsulation, inheritance, and polymorphism in Java.",
			Position:       "Java",
			Difficulty:     "Junior",
			Category:       "OOP",
			Tags:           "OOP,encapsulation,inheritance,polymorphism",
			ExpectedAnswer: "Encapsulation is the process of hiding implementation details. Inheritance allows a class to inherit properties from another class. Polymorphism allows objects to be treated as instances of their parent class.",
		},

		// Go questions
		{
			Title:          "Go Basics: Variables and Data Types",
			Content:        "Please explain the basic data types in Go and how variable declaration differs from other languages.",
			Position:       "Go",
			Difficulty:     "Junior",
			Category:       "Basics",
			Tags:           "variables,data types,basics",
			ExpectedAnswer: "Go has basic data types including int, float, bool, string, etc. Variable declaration uses := for short declaration and var for explicit declaration.",
		},
		{
			Title:          "Go Concurrency",
			Content:        "Please explain goroutines and channels in Go and how they work together.",
			Position:       "Go",
			Difficulty:     "Intermediate",
			Category:       "Concurrency",
			Tags:           "goroutines,channels,concurrency",
			ExpectedAnswer: "Goroutines are lightweight threads managed by the Go runtime. Channels are used for communication between goroutines.",
		},

		// C++ questions (QA or others can use C++)
		{
			Title:          "C++ Basics: Variables and Data Types",
			Content:        "Please explain the basic data types in C++ and their memory sizes.",
			Position:       "QA",
			Difficulty:     "Junior",
			Category:       "Basics",
			Tags:           "variables,data types,basics",
			ExpectedAnswer: "C++ has basic data types including int, float, double, char, bool, etc. The memory size of each type depends on the compiler and system.",
		},
		{
			Title:          "Software Testing Basics",
			Content:        "Please explain the difference between Black Box Testing and White Box Testing.",
			Position:       "QA",
			Difficulty:     "Junior",
			Category:       "Testing",
			Tags:           "testing,black box,white box",
			ExpectedAnswer: "Black Box Testing tests the functionality without knowing the internal structure. White Box Testing tests the internal structure and logic of the code.",
		},
	}

	for _, q := range sampleQuestions {
		var count int64
		// Check if a question with the same title already exists
		db.Model(&model.Question{}).Where("title = ?", q.Title).Count(&count)
		if count == 0 {
			if err := db.Create(&q).Error; err != nil {
				return fmt.Errorf("failed to create sample question: %w", err)
			}
			log.Printf("Created sample question: %s", q.Title)
		}
	}

	return nil
}
