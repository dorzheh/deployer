package main

import (
	"github.com/dorzheh/go-dialog"
	//"time"
	//"strconv"
	"fmt"
)

func main() {
	var d = dialog.New(dialog.CONSOLE, 0)
	//var d = dialog.New(dialog.KDE, 0)
	//var d = dialog.New(dialog.GTK, 0)
	//var d = dialog.New(dialog.X, 0)
	//var d = dialog.New(dialog.AUTO, 0)
	//	d.SetBackTitle("test")
	//	d.SetTitle("test")
	d.SetSize(10, 40)
	//cmd := exec.Command("/bin/bash", "-c", "/home/dorzheh/1")
	//cmd.Run()

	 l := []string{"Selection1", "1", "1", "2046", "1", "10", "20", "0", "0", "Selection2", "2", "1", "0", "2", "10", "20", "0", "0", "Selection3", "3", "1", "0", "3", "10", "20", "0", "0"}
	  d.Mixedform("Title", l[0:]...)

	//strings.Trim(out.String(), "\r\n ")
	//res := d.Calendar(time.Now())
	//fmt.Println(res)

	//res := d.Checklist(0, "Tag 1", "Item 1", "on", "Tag 2", "Item 2", "off", "Tag 3", "Item 3", "off")
	//fmt.Println(res)

	// File select
	//res := d.Fselect("/tmp/test.txt")
	//fmt.Println(res)

	//p := d.Progressbar()
	//time.Sleep(1 * time.Second)
	//for i := 0; i <= 100; i++ {
	//	p.Step(i, "Last: "+strconv.Itoa(100-i)+"%")
	//	time.Sleep(50 * time.Millisecond)
	//}
	//p.Close()

	//d.Infobox("Hello world!")

	//res := d.Inputbox("Hello world!")
	//fmt.Println(res)

	//res := d.Inputmenu(50, "Tag 1", "Item 1", "Tag 2", "Item 2", "Tag 3", "Item 3")
	//fmt.Println(res)

	//res := d.Menu(50, "Tag 1", "Item 1", "Tag 2", "Item 2", "Tag 3", "Item 3")
	//fmt.Println(res)

	//d.Msgbox("Hello world!")

	//res := d.Passwordbox()
	//fmt.Println(res)

	//d.Pause(5)

	//d.Textbox("/tmp/test.txt")

	//res := d.Timebox(time.Now())
	//fmt.Println(res)

	//res := d.Radiolist(0, "Tag 1", "Item 1", "on", "Tag 2", "Item 2", "off", "Tag 3", "Item 3", "off")
	//fmt.Println(res)

	res := d.Yesno()
	fmt.Println(res)

	//res := d.Combobox("Item 1", "Item 2", "Item 3")
	//fmt.Println(res)

	//d.Passivepopup("Hello world!", 5)

	//res := d.Geticon()
	//fmt.Println(res)

	//res := d.Getcolor()
	//fmt.Println(res)

	//res := d.Slider(0, 500, 10)
	//fmt.Println(res)
}
