package web

// Middleware is a type of func which generates a HandleFunc and accepts a next HandleFunc
type Middleware func(next HandleFunc) HandleFunc
