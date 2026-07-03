package store

import (
	"errors"
	"fmt"
	"sync"
)

// Todo represents a single task item in our application.
type Todo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// TodoStore manages an in-memory collection of todos safely across multiple goroutines.
type TodoStore struct {
	mu    sync.RWMutex
	todos map[string]Todo
}

// NewTodoStore initializes and returns a pointer to a new TodoStore.
func NewTodoStore() *TodoStore { // The * in *TodoStore means: "This function doesn't return a heavy struct; it returns a lightweight memory address pointing to where the struct lives."
	return &TodoStore{ // The & in &TodoStore{...} means: "Create this struct in memory right now, and give me its address."
		todos: make(map[string]Todo), // The make(map[string]Todo) means: "Create a new map that can store strings as keys and Todo structs as values."
	}
}

// Add creates a new todo item and stores it safely using a Write Lock.
func (s *TodoStore) Add(title string) Todo {
	s.mu.Lock()         // Acquire exclusive write privileges
	defer s.mu.Unlock() // Release lock when function returns

	// Generate a simple ID based on current length + 1
	id := fmt.Sprintf("%d", len(s.todos)+1)
	
	todo := Todo{
		ID:    id,
		Title: title,
		Done:  false,
	}

	s.todos[id] = todo
	return todo
}

// List returns a slice of all current todos using a Read Lock (multiple readers allowed).
// By using *TodoStore, we ensure that our methods modify the actual instances of our database 
// in memory, rather than making a copy of the whole store structure.
func (s *TodoStore) List() []Todo {
	s.mu.RLock()         // Acquire shared read privileges
	defer s.mu.RUnlock() // Release lock when function returns

	list := make([]Todo, 0, len(s.todos)) //Create a slice that holds Todo dtructs, with 0 length and capacity equal to the number of todos in the store.
	for _, todo := range s.todos {
		list = append(list, todo)
	}
	return list
}

// ToggleComplete switches the done status of a target todo by its ID .
func (s *TodoStore) ToggleComplete(id string) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.todos[id]
	if !exists {
		return Todo{}, errors.New("todo item not found")
	}

	todo.Done = !todo.Done
	s.todos[id] = todo // Save the modified copy back into the map
	return todo, nil
}
