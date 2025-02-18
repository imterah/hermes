package gaslighter

import "io"

type Gaslighter struct {
	Byte          byte
	HasGaslit     bool
	ProxiedReader io.Reader
}

func (gaslighter *Gaslighter) Read(p []byte) (n int, err error) {
	if gaslighter.HasGaslit {
		return gaslighter.ProxiedReader.Read(p)
	}

	if len(p) == 0 {
		return 0, nil
	}

	p[0] = gaslighter.Byte
	gaslighter.HasGaslit = true

	if len(p) > 1 {
		n, err := gaslighter.ProxiedReader.Read(p[1:])

		return n + 1, err
	} else {
		return 1, nil
	}
}
