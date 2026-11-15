package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type titleBlock struct {
	proj    string
	projNum string
	county  string
	str     string
	egr     string
	note    string
}

type data struct {
	rain_i     float64
	lowdrain   int
	max_spread float64
	dia        float64
}

type drain struct {
	//input
	Num int
	W   float64 //inflow width
	L   float64 //segment length
	X   float64 //cross slope
	So  float64 //channel slope
	//output
	Q     float64 //total flow
	H     float64 //water depth
	T     float64 //spread
	Q_in  float64 //in flow
	Q_byp float64 //bypass flow
}

func main() {
	var menu string = "\nOptions:\n  o Open Input File\n  e Edit Input\n  s Save Inupt\n  r Run\n  x Extract from Output file\n  q Quit\nopt> "
	var cmd string
	var iName string
	var drains []drain
	var title titleBlock
	udata := data{7.37, 0, 2.0, 6.0}
	path, _ := os.Getwd()

	for cmd != "q" {
		fmt.Print(menu)
		fmt.Scan(&cmd)
		cmd = strings.ToLower(cmd)

		switch cmd {
		case "o":
			iName, path = getFile(path)
			drains = readInputFile(iName, path, &title, &udata)
		case "e":
			drains = editInput(&title, &udata, drains)
			print_drains(drains)
		case "x":
			iName, path = getFile(path)
			ifile := openFile(iName, path, os.O_RDONLY)
			if ifile == nil {
				continue
			}
			defer ifile.Close()
			drains = extractOutput(ifile, &title, &udata, drains)
		case "s":
			iName = saveInput(&path, &iName, &title, &udata, drains)
			fmt.Printf("New file name: %s\n", iName)
		case "r":
			flowCalc(udata.rain_i, udata.lowdrain, udata.dia, drains)
			writeOuput(&iName, &path, &title, &udata, drains)
		case "q":
			return
		}
	}

	return
}

func flowCalc(ri float64, lowdrain int, dia float64, drains []drain) int {

	var Qby float64

	//forward pass - stop if we hit the low drain
	for i := 0; i < len(drains); i++ {
		Qby = drainFlow(&drains[i], Qby, ri, dia, 0) //pass in 0 as lowdrain so segment flow is added
		if drains[i].Num == lowdrain {
			//fmt.Printf("At low point, Forward Dir: Q=%6.4f\n", drains[i].Q)
			break
		}
	}

	//backward pass - if there's a low drain, start at the end and
	//  work our way back to the low drain
	Qby = 0
	if lowdrain > 0 {
		for i := len(drains) - 1; i >= lowdrain-1 && i >= 0; i-- {
			Qby = drainFlow(&drains[i], Qby, ri, dia, lowdrain) //on the lowdrain the segment flow has already been added
			//fmt.Printf("Backward Dir: Q=%6.4f\n", drains[i].Q)
		}
	}
	return 0
}

// Actual calculation of the flow...
func drainFlow(dr *drain, Qby float64, ri float64, dia float64, lowdrain int) float64 {
	var C float64 = 0.9
	var Y float64 = 0
	var Qweir float64 = 0
	var Qorif float64 = 0

	// fmt.Printf("N=%d  Q=%6.4f  Qby=%6.4f  Add'l=%6.4f\n", dr.Num, dr.Q, Qby, C*ri*dr.W*dr.L/43560)
	if dr.Num == lowdrain {
		dr.Q += Qby //segment flow has already been added in forward loop
	} else {
		dr.Q += Qby + C*ri*dr.W*dr.L/43560
	}
	Y = math.Pow(dr.X, -0.625) * math.Pow(1+math.Sqrt(1+math.Pow(dr.X, 2)), 0.25)
	if dr.So == 0 {
		Hw := math.Pow(dr.Q/(15.75*dia), 1/2.8) / (3.64 * Y)
		Ho := math.Pow(dr.Q/(0.023125*dia*dia), 2) / (3.64 * Y)
		if Hw > Ho {
			dr.H = Hw
		} else {
			dr.H = Ho
		}
	} else {
		dr.H = 0.282 * math.Pow(dr.Q, 0.375) * Y * math.Pow(dr.So, -0.1875)
	}

	dr.T = dr.X * dr.H
	Qweir = weirFlow(dr.So, dia, Y, dr.H)
	Qorif = orificeFlow(dr.So, dia, Y, dr.H)
	if Qweir < Qorif {
		dr.Q_in = Qweir
	} else {
		dr.Q_in = Qorif
	}
	if dr.Q < dr.Q_in {
		dr.Q_in = dr.Q
	}
	dr.Q_byp = dr.Q - dr.Q_in
	Qby = dr.Q_byp

	//fmt.Printf("drain: %v\n", dr)

	return Qby

}

// weir flow type.  interpolate 'So' between 0.002 and 0.05
func weirFlow(So float64, d float64, Y float64, H float64) float64 {
	var Q1 float64 = -1
	var Q2 float64 = -1
	var a float64 = 3.64
	var S1 float64 = 0
	var S2 float64 = 0

	if So <= 0.002 {
		Q1 = 15.75 * d * math.Pow(a*Y*H, 2.8)
		Q2 = 0.505 * d * math.Pow(a*Y*H, 1.5)
		S1 = 0.0
		S2 = 0.002
	} else if So <= 0.01 {
		Q1 = 0.505 * d * math.Pow(a*Y*H, 1.5)
		Q2 = 0.525 * d * math.Pow(a*Y*H, 1.5)
		S1 = 0.002
		S2 = 0.01
	} else if So <= 0.02 {
		Q1 = 0.525 * d * math.Pow(a*Y*H, 1.5)
		Q2 = 1.700 * d * math.Pow(a*Y*H, 1.8)
		S1 = 0.01
		S2 = 0.02
	} else {
		Q1 = 1.700 * d * math.Pow(a*Y*H, 1.8)
		Q2 = 1.050 * d * math.Pow(a*Y*H, 1.5)
		S1 = 0.02
		S2 = 0.05
	}
	return Q1 + (So-S1)/(S2-S1)*(Q2-Q1)
}

// orifice flow type.  interpolate 'So' between 0.002 and 0.05
func orificeFlow(So float64, d float64, Y float64, H float64) float64 {
	var Q1 float64 = -1
	var Q2 float64 = -1
	var a float64 = 3.64
	var S1 float64 = 0
	var S2 float64 = 0

	if So <= 0.002 {
		Q1 = 0.023125 * d * d * math.Pow(a*Y*H, 0.5)
		Q2 = 0.027500 * d * d * math.Pow(a*Y*H, 0.7)
		S1 = 0.0
		S2 = 0.002
	} else if So <= 0.01 {
		Q1 = 0.027500 * d * d * math.Pow(a*Y*H, 0.7)
		Q2 = 0.023750 * d * d * math.Pow(a*Y*H, 0.6)
		S1 = 0.002
		S2 = 0.01
	} else if So <= 0.02 {
		Q1 = 0.023750 * d * d * math.Pow(a*Y*H, 0.6)
		Q2 = 1.024375 * d * d * math.Pow(a*Y*H, 0.6)
		S1 = 0.01
		S2 = 0.02
	} else {
		Q1 = 1.024375 * d * d * math.Pow(a*Y*H, 0.6)
		Q2 = 1.014375 * d * d * math.Pow(a*Y*H, 0.4)
		S1 = 0.02
		S2 = 0.05
	}
	return Q1 + (So-S1)/(S2-S1)*(Q2-Q1)
}

func writeOuput(fName *string, path *string, title *titleBlock, ud *data, drains []drain) {

	var outName string
	{
		n := strings.LastIndexByte(*fName, '.')
		if n >= 0 {
			outName = (*fName)[0:n]
		} else {
			outName = *fName
		}
		outName += ".out"
	}

	sout := fmt.Sprint("\n                     SCUPPER ANALYSIS FOR BRIDGE STRUCTURES\n\n")
	sout += fmt.Sprintln("Input File:  " + *fName)
	sout += fmt.Sprintln("Output File: " + outName)
	sout += headerLine(60, "Project:   ", title.proj, "County: ", title.county)
	sout += headerLine(60, "Project #: ", title.projNum, "By: ", title.egr)
	sout += headerLine(60, "Structure: ", title.str, "Printed: ", "12/31/2026")
	sout += fmt.Sprintln("Note: " + title.note)

	sout += fmt.Sprintln("\n  INPUT DATA :")
	sout += fmt.Sprint("  ------------\n\n")
	sout += fmt.Sprintf("  %5.2f  Rainfall Intensity (in/hr)\n", ud.rain_i)
	sout += fmt.Sprintf("   %2d    Number of Drains\n", len(drains))
	sout += fmt.Sprintf("   %2d    Low point drain number\n", ud.lowdrain)
	sout += fmt.Sprintf("  %5.2f  Allowable spread (ft)\n", ud.max_spread)
	sout += fmt.Sprintf("  %5.2f  Drain diameter (in)\n\n", ud.dia)

	sout += fmt.Sprintln("   DRAIN  SEGMENT  SEGMENT  LONG.  TRANSV.")
	sout += fmt.Sprintln("    NO.    WIDTH   LENGTH   SLOPE   SLOPE")
	sout += fmt.Sprintln("           (ft)     (ft)     (%)     (%)")
	sout += fmt.Sprintln("  ----------------------------------------")
	for _, drn := range drains {
		sout += fmt.Sprintf("    %2d   %6.2f   %6.2f  %6.3f   %6.3f\n", drn.Num,
			drn.W, drn.L, drn.So*100, 100/drn.X)
	}
	sout += fmt.Sprintln("\n  OUTPUT DATA :")
	sout += fmt.Sprint("  -------------\n\n")
	sout += fmt.Sprintln("   DRAIN  SEGMENT   WATER   WATER   DRAIN    BYPASS  DRAIN")
	sout += fmt.Sprintln("    NO.    FLOW     HEIGHT  SPREAD  INFLOW    FLOW    EFF.")
	sout += fmt.Sprintln("           (cfs)     (in)    (ft)   (cfs)     (cfs)   (%) ")
	sout += fmt.Sprint("  --------------------------------------------------------\n")

	var prcEff float64
	var overSpread bool
	for _, drn := range drains {
		if drn.Q != 0 {
			prcEff = drn.Q_in / drn.Q * 100
		}
		sout += fmt.Sprintf("     %2d   %6.4f   %6.3f   %6.3f   %6.4f   %6.4f   %3.0f", drn.Num,
			drn.Q, drn.H*12, drn.T, drn.Q_in, drn.Q_byp, prcEff)
		if drn.T > ud.max_spread {
			sout += fmt.Sprint(" *")
			overSpread = true
		}
		sout += fmt.Sprint("\n")
	}
	if overSpread {
		sout += fmt.Sprintln("      * = Allowable water spread has been exceeded")
	}

	fmt.Print(sout) //Print to screen

	//print to file
	var opt string
	fmt.Printf("\nWrite the above to %s [y/n]: ", outName)
	fmt.Scan(&opt)
	opt = strings.ToLower(opt)
	if opt == "y" {
		fout := openFile(outName, *path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
		if fout == nil {
			return
		}
		defer fout.Close()
		fout.WriteString(sout)
	}

	return
}

func headerLine(wLt int, hdLt string, sLt string, hdRt string, sRt string) string {
	left := fmt.Sprintf("%s%s", hdLt, sLt)
	right := fmt.Sprintf("%s%s", hdRt, sRt)
	n := wLt - len(left)
	for i := 0; i < n; i++ {
		left += " "
	}
	return fmt.Sprintf("%s%s\n", left, right)
}

// used for debugging purposes only
func print_drains(drains []drain) {
	var d drain

	fmt.Println("DEBUG: print_drains()")

	fmt.Println(" N    W     L     X    So      Q       H      T    Q_in   Q_byp")
	for _, d = range drains {
		fmt.Printf("%2d %5.2f %5.2f %5.2f %6.3f %6.4f %6.3f %6.3f %6.3f %6.3f\n",
			d.Num, d.W, d.L, d.X, d.So*100, d.Q, d.H*12, d.T, d.Q_in, d.Q_byp)
	}
}
