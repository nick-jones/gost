package scan

type RunOptions struct {
	stringTableIgnore bool
	stringTableGuess  bool
	noNulls           bool
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

func WithNoNulls() Option {
	return func(o *RunOptions) {
		o.noNulls = true
	}
}
