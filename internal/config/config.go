package config

type Context struct {
	Workers   int
	Level     string
	Component string
	Since     string
	Until     string
	Contains  string
	Format    string
	Output    string
	Buffer    int
}

func NewContext(
	workers int,
	level string,
	component string,
	since string,
	until string,
	contains string,
	format string,
	output string,
	buffer int,
) *Context {
	return &Context{
		Workers:   workers,
		Level:     level,
		Component: component,
		Since:     since,
		Until:     until,
		Contains:  contains,
		Format:    format,
		Output:    output,
		Buffer:    buffer,
	}
}
