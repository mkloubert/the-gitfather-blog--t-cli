// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// the main function / entry point
func main() {
	err := readDotEnvIfAvailable()
	if err != nil {
		// could not read local ENV file

		fmt.Println(err)
		os.Exit(2)
	}

	// we can define the default target language
	// with TGF_DEFAULT_LANGUAGE environment variable
	TGF_DEFAULT_LANGUAGE := strings.TrimSpace(os.Getenv("TGF_DEFAULT_LANGUAGE"))
	if TGF_DEFAULT_LANGUAGE == "" {
		TGF_DEFAULT_LANGUAGE = "english" // this is the fallback
	}

	// storage for: --context / -c
	var context string
	// storage for: --language / -l
	var language string

	// the root command
	var rootCmd = &cobra.Command{
		Use:   "t",
		Short: "Translates text using ChatGPT",
		Long:  `A fast and easy-to-use command line tool to translate texts by The GitFather (https://blog.kloubert.dev/)`,
		Run: func(cmd *cobra.Command, args []string) {
			targetLanguage := strings.TrimSpace(language)
			if targetLanguage == "" {
				targetLanguage = TGF_DEFAULT_LANGUAGE
			}

			textToTranslate := ""

			// first we take all values from non-flag
			// arguments and join them to one string
			// with spaces as separators
			textToTranslate += strings.Join(args, " ")

			// now read text from piped file, if available
			// and add the data as well
			textToTranslate += readFromSTDIN()

			if strings.TrimSpace(textToTranslate) == "" {
				// does not sense to translate empty
				// strings or whitespaces only

				fmt.Println("no valid text to translate")
				os.Exit(3)
			}

			translatedText, err := translateWithGPT(textToTranslate, targetLanguage, context)
			if err != nil {
				fmt.Printf("could not translate text: %v", err)
				os.Exit(4)
			}

			os.Stdout.Write([]byte(translatedText))
		},
	}

	// setup flags ...
	rootCmd.Flags().StringVarP(&language, "context", "c", "", "additional context information for chat model") // --context / -c flag
	rootCmd.Flags().StringVarP(&language, "language", "l", "", "the name of the target language")              // --language / -l flag

	// run the application and parse
	// the command line arguments
	// and other kind of data
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
