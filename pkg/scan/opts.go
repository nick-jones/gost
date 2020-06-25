package scan

type RunOptions struct {
	ignoreStringTable bool
	guessStringTable  bool
}

type Option func(*RunOptions)

func WithStringTableIgnored() Option {
	return func(o *RunOptions) {
		o.ignoreStringTable = true
	}
}

func WithStringTableGuessed() Option {
	return func(o *RunOptions) {
		o.guessStringTable = true
	}
}
