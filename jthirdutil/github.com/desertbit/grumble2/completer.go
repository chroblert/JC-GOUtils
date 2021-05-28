/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2018 Roland Singer [roland.singer@deserbit.com]
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package grumble

import (
	"strings"

	shlex "github.com/desertbit/go-shlex"
)

type completer struct {
	commands *Commands
}

func newCompleter(commands *Commands) *completer {
	return &completer{
		commands: commands,
	}
}

func (c *completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	// Discard anything after the cursor position.
	// This is similar behaviour to shell/bash.
	line = line[:pos]

	var words []string
	if w, err := shlex.Split(string(line), true); err == nil {
		words = w
	} else {
		words = strings.Fields(string(line)) // fallback
	}
	//jlog.Errorf("JCTest:%v,%v,%v_\n",words,pos,string(line))

	prefix := ""
	if len(words) > 0 && pos >= 1 && line[pos-1] != ' ' {
		prefix = words[len(words)-1]
		words = words[:len(words)-1]
	}
	//jlog.Errorf("JCTest:%v,%v_\n",prefix,words)

	// Simple hack to allow auto completion for help.
	if len(words) > 0 && words[0] == "help" {
		words = words[1:]
	}

	var (
		cmds        *Commands
		flags       *Flags
		suggestions [][]rune
	)

	// Find the last commands list.
	if len(words) == 0 {
		cmds = c.commands
	} else
	{
		cmd, rest, err := c.commands.FindCommand(words)
		if err != nil || cmd == nil {
			return
		}

		// Call the custom completer if present.
		if cmd.Completer != nil {
			words = cmd.Completer(prefix, rest)
			for _, w := range words {
				suggestions = append(suggestions, []rune(strings.TrimPrefix(w, prefix)))
			}
			return suggestions, len(prefix)
		}

		// No rest must be there.
		if len(rest) != 0 {
			return
		}

		cmds = &cmd.commands
		flags = &cmd.flags
	}

	if len(prefix) > 0 {
		for _, cmd := range cmds.list {
			if strings.HasPrefix(cmd.FullPath,prefix){
				suggestions = append(suggestions, []rune(strings.TrimPrefix(cmd.FullPath, prefix)))
			}
			if strings.HasPrefix(cmd.Name, prefix) {
				suggestions = append(suggestions, []rune(strings.TrimPrefix(cmd.Name, prefix)))
			}
			for _, a := range cmd.Aliases {
				if strings.HasPrefix(a, prefix) {
					suggestions = append(suggestions, []rune(strings.TrimPrefix(a, prefix)))
				}
			}
		}

		if flags != nil {
			for _, f := range flags.list {
				if len(f.Short) > 0 {
					short := "-" + f.Short
					if len(prefix) < len(short) && strings.HasPrefix(short, prefix) {
						suggestions = append(suggestions, []rune(strings.TrimPrefix(short, prefix)))
					}
				}
				long := "--" + f.Long
				if len(prefix) < len(long) && strings.HasPrefix(long, prefix) {
					suggestions = append(suggestions, []rune(strings.TrimPrefix(long, prefix)))
				}
			}
		}
	} else {
		// [+]210528: 自动填充命令
		//jlog.Errorf("JCTest:words:%v_\n",words[len(words)-1])
		if len(words) > 0 && words[len(words)-1] == "use"{
			//jlog.Errorf("JCTest:cmds.list:%v\n",c.commands.list)
			for _,v := range c.commands.list{
				//jlog.Errorf("%v\n",v.FullPath)
				if v.FullPath != ""{
					suggestions = append(suggestions,[]rune(v.FullPath))
				}
			}
		}
		// end
		for _, cmd := range cmds.list {
			suggestions = append(suggestions, []rune(cmd.Name))
		}
		if flags != nil {
			for _, f := range flags.list {
				suggestions = append(suggestions, []rune("--"+f.Long))
				if len(f.Short) > 0 {
					suggestions = append(suggestions, []rune("-"+f.Short))
				}
			}
		}
	}
	//if prefix == "use"{
	//	for _,v := range cmds.list{
	//		jlog.Errorf("%v\n",v.FullPath)
	//		suggestions = append(suggestions,[]rune(v.FullPath))
	//	}
	//}
	// Append an empty space to each suggestions.
	for i, s := range suggestions {
		suggestions[i] = append(s, ' ')
	}
	//for _,v := range suggestions{
	//	jlog.Errorf("JCTest:%v,%v_\n",string(v),prefix)
	//
	//}

	return suggestions, len(prefix)
}
