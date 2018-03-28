/*
 * Copyright (C) 2017-2018 Alibaba Group Holding Limited
 */
package cli

import (
	"strings"
)

type FlagDetector interface {
	DetectFlag(name string, isShorthand bool) (*Flag, error)
}

type Parser struct {
	current int
	args    []string
	detector FlagDetector
}

func NewParser(args []string, detector FlagDetector) *Parser {
	return &Parser {
		args:    args,
		current: 0,
		detector: detector,
	}
}

func (p *Parser) ReadNextArg() (arg string, more bool, err error) {
	for {
		arg, _, more, err = p.readNext()
		if err != nil {
			return
		}
		if !more {
			return
		}
		if arg != "" {
			return
		}
	}
}

func (p *Parser) GetRemains() []string {
	return p.args[p.current:]
}

func (p *Parser) ReadAll() ([]string, error) {
	r := make([]string, 0)
	for {
		arg, _, more, err := p.readNext()
		if err != nil {
			return r, err
		}
		if arg != "" {
			r = append(r, arg)
		}
		if !more {
			return r, nil
		}
	}
}

func (p *Parser) readNext() (arg string, flag *Flag, more bool, err error) {
	if p.current >= len(p.args) {
		more = false
		return
	}
	s := p.args[p.current]
	p.current++
	more = true

	var prefixLen int
	if strings.HasPrefix(s, "--") {
		prefixLen = 2
	} else if strings.HasPrefix(s, "-") {
		prefixLen = 1
	} else {
		prefixLen = 0
	}

	if prefixLen > 0 {
		if name, value, ok := SplitWith(s[prefixLen:], "=:"); ok {
			flag, err = p.detector.DetectFlag(name, prefixLen == 1)
			if err != nil {
				return
			}
			err = flag.putValue(value)
			if err != nil {
				return
			}
			return
		} else {
			flag, err = p.detector.DetectFlag(name, prefixLen == 1)
			if err != nil {
				return
			}
			switch flag.AssignedMode {
			case AssignedNone:
				flag.putValue("")
			case AssignedDefault:
				if value, ok := p.readNextValue(false); ok {
					flag.putValue(value)
				} else {
					flag.putValue("")
				}
			case AssignedOnce:
				if value, ok := p.readNextValue(true); ok {
					flag.putValue(value)
				}
			case AssignedRepeatable:
				if value, ok := p.readNextValue(true); ok {
					flag.putValue(value)
				}
			}
			return
		}
	} else {
		arg = s
		return
	}
}

//
//
func (p *Parser) readNextValue(force bool) (string, bool) {
	if p.current >= len(p.args) {
		return "", false
	}
	s := p.args[p.current]
	if force {
		p.current++
		return s, true
	} else {
		if !strings.HasPrefix(s, "--") && !strings.HasPrefix(s, "-") {
			p.current++
			return s, true
		} else {
			return "", false
		}
	}
}