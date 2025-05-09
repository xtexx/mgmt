// Mgmt
// Copyright (C) James Shubin and the project contributors
// Written by James Shubin <james@shubin.ca> and the project contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// Additional permission under GNU GPL version 3 section 7
//
// If you modify this program, or any covered work, by linking or combining it
// with embedded mcl code and modules (and that the embedded mcl code and
// modules which link with this program, contain a copy of their source code in
// the authoritative form) containing parts covered by the terms of any other
// license, the licensors of this program grant you additional permission to
// convey the resulting work. Furthermore, the licensors of this program grant
// the original author, James Shubin, additional permission to update this
// additional permission if he deems it necessary to achieve the goals of this
// additional permission.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type function struct {
	// MclName is the name of the package of the function in mcl.
	MgmtPackage string `yaml:"mgmtPackage"`
	// MclName is the name of the function in mcl.
	MclName string `yaml:"mclName"`
	// InternalName is the name used inside the templated file.
	// Used to avoid clash between same functions from different packages.
	InternalName string `yaml:"internalName"`
	// Help is the docstring of the function, including // and
	// new lines.
	Help string `yaml:"help"`
	// GolangPackage is the representation of the package.
	GolangPackage *golangPackage `yaml:"golangPackage"`
	// GolangFunc is the name of the function in golang.
	GolangFunc string `yaml:"golangFunc"`
	// Errorful indicates whether the golang function can return an error
	// as second argument.
	Errorful bool `yaml:"errorful"`
	// Args is the list of the arguments of the function.
	Args []arg `yaml:"args"`
	// ExtraGolangArgs are arguments that are added at the end of the go call.
	// e.g. strconv.ParseFloat("3.1415", 64) could require add 64.
	ExtraGolangArgs []arg `yaml:"extraGolangArgs"`
	// Return is the list of arguments returned by the function.
	Return []arg `yaml:"return"`
}

func parseFuncs(c config, f functions, path, templates string) error {
	templateFiles, err := filepath.Glob(templates)
	if err != nil {
		return err
	}
	for _, tpl := range templateFiles {
		log.Printf("Template: %s", tpl)
		err = generateTemplate(c, f, path, tpl, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func generateTemplate(c config, f functions, path, templateFile, finalName string) error {
	log.Printf("Reading: %s", templateFile)
	basename := filepath.Base(templateFile)
	tplFile, err := os.ReadFile(templateFile)
	if err != nil {
		return err
	}
	t, err := template.New(basename).Parse(string(tplFile))
	if err != nil {
		return err
	}
	if finalName == "" {
		finalName = strings.TrimSuffix(basename, ".tpl")
	}
	finalPath := filepath.Join(path, finalName)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	return t.Execute(finalFile, struct {
		Packages  golangPackages
		Functions []function
	}{
		c.Packages,
		f,
	})
}

// MakeGolangArgs translates the func args to golang args.
func (obj *function) MakeGolangArgs() (string, error) {
	var args []string
	for i, a := range obj.Args {
		gol, err := a.ToGolang(fmt.Sprintf("args[%d]", i))
		if err != nil {
			return "", err
		}

		switch a.Type {
		case "int":
			gol = fmt.Sprintf("int(%s)", gol)
		case "[]byte":
			gol = fmt.Sprintf("[]byte(%s)", gol)
		}
		args = append(args, gol)
	}
	for _, a := range obj.ExtraGolangArgs {
		args = append(args, a.Value)
	}
	return strings.Join(args, ", "), nil
}

// Signature generates the mcl signature of the function.
func (obj *function) Signature() (string, error) {
	var args []string
	for _, a := range obj.Args {
		mcl, err := a.ToMcl()
		if err != nil {
			return "", err
		}
		args = append(args, mcl)
	}
	var returns []string
	for _, a := range obj.Return {
		mcl, err := a.ToMcl()
		if err != nil {
			return "", err
		}
		returns = append(returns, mcl)
	}
	return fmt.Sprintf("func(%s) %s", strings.Join(args, ", "), returns[0]), nil
}

// MakeGoReturn returns the golang signature of the return.
func (obj *function) MakeGoReturn() (string, error) {
	return obj.Return[0].OldToGolang()
}

// ConvertStart returns the start of a casting function required to convert from
// mcl to golang.
func (obj *function) ConvertStart() string {
	t := obj.Return[0].Type
	switch t {
	case "int":
		return "int64("
	case "[]byte":
		return "string("
	default:
		return ""
	}
}

// ConvertStop returns the end of the conversion function required to convert
// from mcl to golang.
func (obj *function) ConvertStop() string {
	t := obj.Return[0].Type
	switch t {
	case "int", "[]byte":
		return ")"
	default:
		return ""
	}
}

// MakeGolangTypeReturn returns the mcl signature of the return.
func (obj *function) MakeGolangTypeReturn() string {
	t := obj.Return[0].Type
	switch t {
	case "int64":
		t = "int"
	case "float64":
		t = "float"
	}
	return t
}
