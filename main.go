package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func main() {
	app := cli.NewApp()
	app.Name = "gah"
	app.Usage = "This is AtCoder helper."
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:    "setup",
			Aliases: []string{"s"},
			Usage:   "setup",
			Action:  setupAction,
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "test",
			Action:  testAction,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.After = func(c *cli.Context) error {
		return nil
	}

	app.Run(os.Args)
}

func setupAction(c *cli.Context) {
	if len(c.Args()) < 1 {
		fmt.Printf("please specify contest name.\n")
		return
	}
	var contestName = c.Args().First()

	var contestUrl = "https://" + contestName + ".contest.atcoder.jp/"
	var assignmentsUrl = contestUrl + "assignments/"

	doc, err := goquery.NewDocument(assignmentsUrl)
	if err != nil {
		fmt.Printf("failed to fetch %s\n", assignmentsUrl)
		return
	}

	var titles []string
	var urls []string
	doc.Find("#outer-inner > table > tbody > tr > td.center > a").Each(func(_ int, s *goquery.Selection) {
		title := s.Text()
		url, _ := s.Attr("href")
		titles = append(titles, title)
		urls = append(urls, url)
	})

	var projectPath = os.Getenv("HOME") + "/Work/atcoder/"
	var contestPath = projectPath + contestName + "/"
	var testPath = contestPath + "test/"

	os.MkdirAll(contestPath, 0755)
	fmt.Printf("mkdir %s\n", contestPath)

	os.MkdirAll(testPath, 0755)
	fmt.Printf("mkdir %s\n", testPath)

	fmt.Println("")

	for i := range titles {
		var templatePath = projectPath + "template.cpp"
		src, _ := os.Open(templatePath)
		defer src.Close()

		var dstName = contestPath + titles[i] + ".cpp"
		_, err := os.Stat(dstName)
		if err != nil {
			dst, _ := os.Create(dstName)
			defer dst.Close()

			_, _ = io.Copy(dst, src)
			fmt.Printf("create %s\n", dstName)
		}

		var taskUrl = contestUrl + urls[i]
		taskDoc, err := goquery.NewDocument(taskUrl)
		if err != nil {
			fmt.Printf("failed to fetch %s\n", taskUrl)
			return
		}
		fmt.Printf("fetch %s\n", taskUrl)

		var index = 1
		taskDoc.Find("#task-statement > span > span.lang-en > div > section > pre").Each(func(_ int, s *goquery.Selection) {
			text := s.Text()

			var testFileName = titles[i] + "_" + fmt.Sprint((index+1)/2)
			if index%2 == 1 {
				testFileName += ".in"
			} else {
				testFileName += ".out"
			}
			testFile, _ := os.Create(testPath + testFileName)
			defer testFile.Close()
			testFile.Write(([]byte)(text))

			fmt.Printf("create %s\n", testFileName)

			index++
		})
		fmt.Println("")
	}

	fmt.Printf("finished\n")
}

func testAction(c *cli.Context) {
	var testPath = "test/"
	if len(c.Args()) < 1 {
		fmt.Printf("please specify task name.\n")
		return
	}
	var taskName = c.Args().First()

	// compile
	var srcFileName = taskName + ".cpp"
	fmt.Printf("compile... %s\n", srcFileName)
	err := exec.Command("g++", srcFileName).Run()
	if err != nil {
		fmt.Println("compile error")
		return
	}
	fmt.Printf("finished\n\n")

	// exec test
	var index = 1
	for {
		var testName = taskName + "_" + fmt.Sprint(index)
		var inputFileName = testPath + testName + ".in"
		_, err := os.Stat(inputFileName)
		if err != nil {
			// do not exist file
			break
		}

		input, _ := ioutil.ReadFile(inputFileName)
		cmd := exec.Command("./a.out")
		stdin, _ := cmd.StdinPipe()
		io.WriteString(stdin, string(input))
		stdin.Close()
		got, _ := cmd.Output()

		var outputFileName = testPath + testName + ".out"
		expected, _ := ioutil.ReadFile(outputFileName)

		fmt.Printf("%s\n", testName)
		if string(got) == string(expected) {
			fmt.Println("\x1b[32mAC\x1b[0m")
		} else {
			fmt.Println("\x1b[31mWA\x1b[0m")
		}
		fmt.Printf("expected: %s", string(expected))
		fmt.Printf("got: %s", string(got))

		index++
		fmt.Println("")
	}
}
