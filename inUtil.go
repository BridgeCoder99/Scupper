package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func saveInput(path *string, iName *string, title *titleBlock, ud *data, drains []drain) string {

	var fName string
	if len(*iName) == 0 {
		fName, *path = getFile(*path)
	} else {
		n := strings.LastIndexByte(*iName, '.')
		if n >= 0 {
			fName = (*iName)[0:n] + ".in"
		} else {
			fName = *iName + ".in"
		}
		fmt.Printf("Save to this file: %s [y/n]: ", fName)
		var ans string
		fmt.Scan(&ans)
		ans = strings.ToLower(ans)
		if ans != "y" {
			fName, *path = getFile(*path)
		}
	}

	fout := openFile(fName, *path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if fout == nil {
		return fName
	}
	defer fout.Close()

	s := "<<TITLE>>\n"
	s += fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n", title.proj, title.projNum,
		title.county, title.str, title.egr, title.note)
	s += "<<END TITLE>>\n"
	s += "<<Intensity(in/hr), Low Point, Allow Spread(ft), Drain Dia(in)>>\n"
	s += fmt.Sprintf("%f,%d,%f,%f\n", ud.rain_i, ud.lowdrain, ud.max_spread, ud.dia)
	s += "<<DRAINS>>\n"
	s += "<<Number, Width(ft), Length(ft), X Slope(h/v), Slope(v/h)>>\n"
	for _, drn := range drains {
		s += fmt.Sprintf("%d,%f,%f,%f,%f\n", drn.Num, drn.W, drn.L, drn.X, drn.So)
	}
	s += "<<END DRAINS>>"

	fout.WriteString(s)

	return fName
}

func readInputFile(fName string, path string, title *titleBlock, udata *data) []drain {

	var drains []drain
	ifile := openFile(fName, path, os.O_RDONLY)
	if ifile == nil {
		return drains
	}
	defer ifile.Close()
	ioRead := bufio.NewReader(ifile)

	//read each line of the file until we hit <<END DRAINS>> or EOF
	for s, er := ioRead.ReadString('\n'); !strings.Contains(s, "<<END DRAINS>>"); s, er = ioRead.ReadString('\n') {
		if er != nil {
			if er == io.EOF {
				return drains
			} else {
				fmt.Println(er)
			}
		}
		//read in the title lines
		if strings.Contains(s, "<<TITLE") {
			readTitleBlock(ioRead, title)
			//read in the Rain Intensity line
		} else if strings.Contains(s, "<<Intensity") {
			//read the next line which is where the data is
			s, er = ioRead.ReadString('\n')
			sl := strings.Split(s, ",")
			if len(sl) >= 4 {
				udata.rain_i = parseFloat(sl[0])
				udata.lowdrain = parseInt(sl[1])
				udata.max_spread = parseFloat(sl[2])
				udata.dia = parseFloat(sl[3])
			}
			//read in the drains table
		} else if strings.Contains(s, "<<DRAINS>>") {
			drains = readDrains(ioRead)
		}
	}
	return drains
}

func readTitleBlock(ioRead *bufio.Reader, title *titleBlock) {
	var end bool
	title.proj, end = readTitleLine(ioRead, "<<END TITLE>>")
	if end {
		return
	}
	title.projNum, end = readTitleLine(ioRead, "<<END TITLE>>")
	if end {
		return
	}
	title.county, end = readTitleLine(ioRead, "<<END TITLE>>")
	if end {
		return
	}
	title.str, end = readTitleLine(ioRead, "<<END TITLE>>")
	if end {
		return
	}
	title.egr, end = readTitleLine(ioRead, "<<END TITLE>>")
	if end {
		return
	}
	title.note, end = readTitleLine(ioRead, "<<END TITLE>>")
}

func readTitleLine(ioRead *bufio.Reader, until string) (string, bool) {
	s, err := ioRead.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return s, true
		}
		fmt.Println("readTitleLine(): " + err.Error())
	}
	s = strings.TrimRight(s, "\r\n")
	if strings.Contains(s, until) {
		s = ""
		return s, true
	}
	return s, false
}

func readDrains(ioRead *bufio.Reader) []drain {
	var drains []drain

	for ln, err := ioRead.ReadString('\n'); !strings.Contains(ln, "<<END"); ln, err = ioRead.ReadString('\n') {
		if strings.Contains(ln, "<<Number") {
			continue
		}
		var drn drain
		sl := strings.Split(ln, ",")
		if len(sl) >= 5 {
			drn.Num = parseInt(sl[0])
			drn.W = parseFloat(sl[1])
			drn.L = parseFloat(sl[2])
			drn.X = parseFloat(sl[3])
			drn.So = parseFloat(sl[4])
			drains = append(drains, drn)
		}
		if err != nil {
			if err == io.EOF {
				return drains
			}
			fmt.Println("Error reading drain data: " + err.Error())
			return drains
		}
	}
	return drains
}

func editInput(title *titleBlock, ud *data, drains []drain) []drain {

	var cmd string
	var itm int

	fmt.Println("                            <<Scupper Edit Utility>>")
	fmt.Println("Commands\n\tLn#: Edit Line #\n\t  p: Print Input\n\t  a: Append Drain")
	fmt.Print("\t  i: Insert Drain\n\t  d: Delete Drain\n\t  q: Quit editor\n\n")

	// rainLn = line # that the rain intensity is displayed on
	// endTbl = last line # of the drains table
	rainLn, endTbl := printInput(title, ud, drains)

	for !strings.Contains(cmd, "q") {
		//print the input data

		fmt.Print("\nLn# p a i d q> ")
		fmt.Scan(&cmd)

		if isInt(cmd) {
			itm = parseInt(cmd)
			if itm < rainLn {
				switch itm {
				case 1:
					title.proj = getUserText("Project: ", title.proj)
				case 2:
					title.projNum = getUserText("Project #: ", title.projNum)
				case 3:
					title.county = getUserText("County: ", title.county)
				case 4:
					title.str = getUserText("Structure #: ", title.str)
				case 5:
					title.egr = getUserText("By: ", title.egr)
				case 6:
					title.note = getUserText("Note: ", title.note)
				}
			} else if itm == rainLn {
				ud.rain_i = getUserFloat("Rain Intensity (in/hr)", ud.rain_i)
				newlow := getUserFloat("Low Drain", float64(ud.lowdrain))
				ud.lowdrain = int(newlow)
				ud.max_spread = getUserFloat("Allowable Spread", ud.max_spread)
				ud.dia = getUserFloat("Drain Diameter (in)", ud.dia)
			} else if itm > rainLn && itm <= endTbl {
				idx := itm - rainLn - 1
				getDrainInput(&drains[idx])
			}
		} else if strings.Contains(cmd, "p") {
			rainLn, endTbl = printInput(title, ud, drains)
		} else if strings.Contains(cmd, "a") {
			drains = appendDrain(drains)
			endTbl++
		} else if strings.Contains(cmd, "i") {
			drains = insertDrain(drains, rainLn+1)
			endTbl++
		} else if strings.Contains(cmd, "d") {
			drains = deleteDrain(drains, rainLn+1)
			endTbl--
		}

	}
	return drains
}

func printInput(title *titleBlock, ud *data, drains []drain) (int, int) {
	var rainLn, endTbl int

	ln := 1
	fmt.Println("  Line #  Title Block Text")
	fmt.Printf("    %2d  Project: %s\n", ln, title.proj)
	ln++
	fmt.Printf("    %2d  Project#: %s\n", ln, title.projNum)
	ln++
	fmt.Printf("    %2d  County: %s\n", ln, title.county)
	ln++
	fmt.Printf("    %2d  Structure #: %s\n", ln, title.str)
	ln++
	fmt.Printf("    %2d  By: %s\n", ln, title.egr)
	ln++
	fmt.Printf("    %2d  Note: %s\n", ln, title.note)
	ln++

	fmt.Println("\n  Line #  Intensity(in/hr)  Low Point  Allow Spread(ft)  Drain Dia(in)")
	fmt.Printf("    %2d          %5.2f           %2d          %5.2f            %5.2f\n\n",
		ln, ud.rain_i, ud.lowdrain, ud.max_spread, ud.dia)
	rainLn = ln
	ln++

	fmt.Println("  Line #  Width(ft)  Length(ft)  X Slope(h/v)  Slope(v/h)")
	for _, drn := range drains {
		fmt.Printf("    %2d      %5.2f       %5.2f        %5.2f      %7.4f\n", ln,
			drn.W, drn.L, drn.X, drn.So)
		ln++
	}
	endTbl = ln - 1

	return rainLn, endTbl
}

func getUserText(prompt string, cur string) string {
	if len(cur) > 0 {
		fmt.Println("Replace: " + cur)
	}
	fmt.Print(prompt)
	var newln string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		newln = scanner.Text()
	}
	if len(newln) <= 0 {
		return cur
	}
	return newln
}

func getUserFloat(prompt string, dflt float64) float64 {
	var txt string
	fmt.Printf("%s Default = %8.4f:  ", prompt, dflt)
	_, err := fmt.Scanln(&txt)
	if err != nil {
		if !strings.Contains(err.Error(), "newline") {
			fmt.Println(err.Error())
		}
		return dflt
	}
	if len(txt) <= 0 {
		return dflt
	}
	return parseFloat(txt)
}

func getDrainInput(drn *drain) drain {
	drn.W = getUserFloat("Segment Width (ft) ", drn.W)
	drn.L = getUserFloat("Segment Length (ft)", drn.L)
	drn.X = getUserFloat("Cross Slope (h/v)  ", drn.X)
	drn.So = getUserFloat("Long. Slope (%)    ", drn.So*100) / 100
	return *drn
}

func appendDrain(drains []drain) []drain {
	var d drain
	d = getDrainInput(&d)
	drains = append(drains, d)
	drains = numberDrains(drains)
	return drains
}
func insertDrain(drains []drain, bgnln int) []drain {
	nof := len(drains)
	if nof == 0 {
		return appendDrain(drains)
	}

	var s string
	fmt.Print("Line # to insert drain before: ")
	fmt.Scan(&s)
	if isInt(s) {
		ln := parseInt(s)
		if ln >= bgnln && ln < bgnln+nof {
			var drn drain
			drn = getDrainInput(&drn)
			drains = slices.Insert(drains, ln-bgnln, drn)
		} else {
			fmt.Printf("Line # must be between %d and %d\n", bgnln, bgnln+nof-1)
		}
	} else {
		fmt.Println("Looking for a line #")
	}
	drains = numberDrains(drains)

	return drains
}

func deleteDrain(drains []drain, bgnln int) []drain {
	nof := len(drains)
	if nof == 0 {
		return drains
	}

	var s string
	fmt.Print("Line # of drain to delete: ")
	fmt.Scan(&s)
	if isInt(s) {
		ln := parseInt(s)
		if ln >= bgnln && ln < bgnln+nof {
			drains = slices.Delete(drains, ln-bgnln, ln-bgnln+1)
		} else {
			fmt.Printf("Line # must be between %d and %d\n", bgnln, bgnln+nof-1)
		}
	} else {
		fmt.Println("Looking for a line #")
	}
	drains = numberDrains(drains)

	return drains
}

func numberDrains(drains []drain) []drain {
	for i, drn := range drains {
		drn.Num = i + 1
	}
	return drains
}

func extractOutput(ifile *os.File, title *titleBlock, ud *data, drains []drain) []drain {

	pread := bufio.NewReader(ifile)

	//Top lines before "INPUT DATA" are the "title"
	// Read until we hit a blank line
	var s string
	var er error
	for s, er = pread.ReadString('\n'); !strings.Contains(s, "INPUT DATA"); s, er = pread.ReadString('\n') {
		if er != nil {
			fmt.Println("extractInput(): " + er.Error())
			return drains
		}
		if strings.Contains(s, "Project:") {
			title.proj, _ = strings.CutPrefix(s, "Project:")
		} else if strings.Contains(s, "Project #") {
			title.projNum, _ = strings.CutPrefix(s, "Project #:")
		} else if strings.Contains(s, "County:") {
			title.county, _ = strings.CutPrefix(s, "County:")
		} else if strings.Contains(s, "Structure:") {
			title.str, _ = strings.CutPrefix(s, "Structure:")
		} else if strings.Contains(s, "Engineer:") {
			title.egr, _ = strings.CutPrefix(s, "Engineer:")
		} else if strings.Contains(s, "Note:") {
			title.note, _ = strings.CutPrefix(s, "Note:")
		}
	}

	//Read the "Input Data" section until we hit the Drain table
	var sl []string
	for !strings.Contains(s, "DRAIN") /* || !strings.Contains(s, "OUTPUT") */ {
		if er != nil {
			if er == io.EOF {
				break
			}
			fmt.Println("2 Error reading file: " + er.Error())
			os.Exit(1)
		}
		if strings.Contains(s, "Intensity") {
			sl = strings.Split(strings.TrimLeft(s, " \t"), " ")
			ud.rain_i = parseFloat(sl[0])
		} else if strings.Contains(s, "Region") {
			sl = strings.Split(strings.TrimLeft(s, " \t"), " ")
			rgn := parseInt(sl[0])
			switch rgn {
			case 1:
				ud.rain_i = 6.88
			case 2:
				ud.rain_i = 7.41
			default:
				ud.rain_i = 7.37
			}
		} else if strings.Contains(s, "Low point") {
			sl = strings.Split(strings.TrimLeft(s, " \t"), " ")
			ud.lowdrain = parseInt(sl[0])
		} else if strings.Contains(s, "spread") {
			sl = strings.Split(strings.TrimLeft(s, " \t"), " ")
			ud.max_spread = parseFloat(sl[0])
		} else if strings.Contains(s, "diameter") {
			sl = strings.Split(strings.TrimLeft(s, " \t"), " ")
			ud.dia = parseFloat(sl[0])
		}
		s, er = pread.ReadString('\n')
	}

	//Read in the drain data table
	// The top line of the table headings has been read in. skip 3 more lines to
	// get to the first line of data
	for i := 0; i < 3; i++ {
		if er != nil {
			fmt.Println("3 Error reading file: " + er.Error())
			os.Exit(1)
		}
		s, er = pread.ReadString('\n')
	}

	for s, er = pread.ReadString('\n'); !strings.Contains(s, "OUTPUT"); {
		if er != nil {
			fmt.Println("4 Error reading file: " + er.Error())
			os.Exit(1)
		}

		if !isBlankLine(s) {
			scanner := bufio.NewScanner(strings.NewReader(s))
			scanner.Split(bufio.ScanWords)
			var drn drain
			col := 0
			for scanner.Scan() {
				sl := (scanner.Text())
				switch col {
				case 0:
					drn.Num = parseInt(sl)
				case 1:
					drn.W = parseFloat(sl)
				case 2:
					drn.L = parseFloat(sl)
				case 3:
					drn.So = parseFloat(sl) / 100.0
				case 4:
					drn.X = 100.0 / parseFloat(sl)
				}
				col++
			}
			drains = append(drains, drn)
		}
		s, er = pread.ReadString('\n')
	}

	return drains
}

// read each character of "s"; if every character is either a blank, tab,
// carriage return or new line then return true
func isBlankLine(s string) bool {
	t := strings.Trim(s, "\r\n\t ")

	if len(t) <= 1 {
		return true
	}
	return false
}

func parseFloat(s string) float64 {
	f, er := strconv.ParseFloat(strings.Trim(s, " \t\n"), 64)
	if er != nil {
		fmt.Println("func parseFloat(): " + er.Error())
		return 0.0
	}
	return f
}

func parseInt(s string) int {
	j, er := strconv.Atoi(s)
	if er != nil {
		fmt.Println("func parseInt(): " + er.Error())
		return 0
	}
	return j
}
