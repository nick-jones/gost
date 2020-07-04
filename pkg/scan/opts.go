package scan

type RunOptions struct {
	stringTableIgnore bool
	stringTableGuess  bool
	permitNulls       bool
}

type Option func(*RunOptions)

func WithStringTableIgnored() Option {
	return func(o *RunOptions) {
		o.stringTableIgnore = true
	}
}

func WithStringTableGuessed() Option {
	return func(o *RunOptions) {
		o.stringTableGuess = true
	}
}

func WithNullsPermitted() Option {
	return func(o *RunOptions) {
		o.permitNulls = true
	}
}
