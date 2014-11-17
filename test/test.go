// Note: this illustrates ValuesForKey() and ValuesForPath() methods

package main

import (
	"fmt"
	"github.com/clbanning/mxj"
	"log"
)

var xmldata = []byte(`
   <books>
      <book>
         <author>William H. Gaddis</author>
         <title>20</title>
         <review>One of the great seminal American novels of the 20th century.</review>
      </book>
      <book>
         <author>Austin Tappan Wright</author>
         <title>30</title>
         <review>An example of earlier 20th century American utopian fiction.</review>
      </book>
      <book>
         <author>John Hawkes</author>
         <title>10</title>
         <review>A lyrical novel about the construction of Ft. Peck Dam in Montana.</review>
      </book>
      <book>
         <author>T.E. Porter</author>
         <title>40</title>
         <review>A magical novella.</review>
      </book>
   </books>
`)

func main() {
	//	fmt.Println(string(xmldata))

	m, err := mxj.NewMapXml(xmldata)
	if err != nil {
		log.Fatal("err:", err.Error())
	}

	v, _ := m.ValuesForPath("books.book")
	for _, i := range v {
		ch := i.(map[string]interface{})
		val := ch["title"].(int)
		if val == 10 {
			fmt.Printf("\t%+v\n", ch["review"])
		}
	}

	// v, _ = m.ValuesForPath("books.book")
	// fmt.Println("path: books.book; len(v):", len(v))
	// for _, vv := range v {
	// 	fmt.Printf("\t%+v\n", vv)
	// }

	// v, _ = m.ValuesForPath("books.*")
	// fmt.Println("path: books.*; len(v):", len(v))
	// for _, vv := range v {
	// 	fmt.Printf("\t%+v\n", vv)
	// }

	// v, _ = m.ValuesForPath("books.*.title")
	// fmt.Println("path: books.*.title len(v):", len(v))
	// for _, vv := range v {
	// 	fmt.Printf("\t%+v\n", vv)
	// }

	// v, _ = m.ValuesForPath("books.*.*")
	// fmt.Println("path: books.*.*; len(v):", len(v))
	// for _, vv := range v {
	// 	fmt.Printf("\t%+v\n", vv)
	// }
}
