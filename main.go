package main
import (
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/opesun/goquery"
//	"net"
	"time"
	"net/url"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"bufio"
	"strings"
	"flag"
)
var (
	Dir = flag.String("d","dir","dir1")
	NewFile = flag.String("n","newData.csv","dir2")
	Day = flag.Int("p",30,"data count")
	Init = flag.Bool("init",false,"init")
)
type InfoData struct {
	name string
	url string
	Time string
}

func (self *InfoData) IsSame(d *InfoData) bool {
	if self.name != d.name {
		return false
	}
	if self.url != d.url {
		return false
	}
	if self.Time != d.Time {
		return false
	}
	return true
}
func (self *InfoData) SaveData(dir string) {
	os.MkdirAll(dir,0777)
	f,err :=os.OpenFile(filepath.Join(dir,self.Time+".csv"),os.O_CREATE|os.O_APPEND|os.O_RDWR,0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_,err = f.WriteString(fmt.Sprintf("%s  %s\n",self.name,self.url))
	if err != nil {
		panic(err)
	}
}
type SiteInfo struct {

	Client *http.Client
	Header http.Header
//	Host string
	OldData []*InfoData
	NewData []*InfoData
	Count int

}

func (self *SiteInfo) ReadOldData(dir string) error {
	now := time.Now()

	return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		t := strings.Split(fi.Name(),".")[0]
		tis ,err := time.Parse("2006-01-02",t)
		if err != nil {
			panic(err)
		}
		newTime := tis.Add(time.Hour*24*time.Duration(*Day))
		if newTime.Before(now) {
			return nil
		}
		fmt.Println(tis,path)
//		if self.Count >1000 {
//			return nil
//		}
		f,err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		rd := bufio.NewReader(f)
		//count :=0
		for {
			line,err := rd.ReadString('\n')
			if err != nil {
				if io.EOF == err {
					break
				}
			}
//			fmt.Println(line[:len(line)-1])
			ls := strings.Split(line[:len(line)-1],"  ")
//			self.OldData[us[len(us)-1]]= &InfoData{ls[0],ls[1],t}
			self.OldData = append(self.OldData,&InfoData{ls[0],ls[1],t})
		//	count++
			self.Count ++
			if self.Count!=len(self.OldData){
//				fmt.Println(line)
				self.Count = len(self.OldData)
			}
		}
		//fmt.Println(count)
		return nil
	})

}

func (self *SiteInfo)Init(dir string) error {

//	self.Host = Host
	self.Count = 0
//	self.OldData = make(map[string]*InfoData)
	self.Header = make(http.Header)
	self.Header.Add("Accept","text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	self.Header.Add("Connection","keep-alive")
	self.Header.Add("Accept-Encoding","gzip, deflate, sdch")
	self.Header.Add("Accept-Language","zh-CN,zh;q=0.8")
	self.Header.Add("User-Agent","Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/58.0.3029.110 Chrome/58.0.3029.110 Safari/537.36")

	self.Client = new(http.Client)
	return self.ReadOldData(dir)
}
func (self *SiteInfo) ClientDO (path string) ([]byte,error) {

	Req,err := http.NewRequest("GET",path,nil)
	if err != nil {
		return nil,err
	}
	Req.Header = self.Header
	res,err := self.Client.Do(Req)
	if err != nil {
		return nil,err
	}
	defer res.Body.Close()
//	fmt.Println(res.StatusCode)
	if res.StatusCode != 200 {
		b,err:=ioutil.ReadAll(res.Body)
		fmt.Println(string(b),err)
		return nil,fmt.Errorf("%d",res.StatusCode)
	}

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(res.Body)
		defer reader.Close()
	default:
		reader = res.Body
	}
	return ioutil.ReadAll(reader)

}

func (self *SiteInfo) SaveNewData(NewFile string) error {
	le:= len(self.NewData)
	fmt.Println("new",le)
	if le == 0 {
		return nil
	}

	f,err :=os.OpenFile(NewFile,os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_SYNC,0777)
	if err != nil {
		return err
	}
	defer f.Close()
	for _,d := range self.NewData {
		fmt.Println(d)
		_,err = f.WriteString(d.Time+"  "+d.name+"  "+d.url+"\n")
		if err != nil {
			return err
		}
	}
	return nil

}

func (self *SiteInfo)FindSameOld(d *InfoData) bool{
	le := len(self.OldData)
	if le == 0 {
		return false
	}
	for i:= le-1;i>=0;i--{
		if self.OldData[i].IsSame(d) {
			return true
		}
	}
	return false
}
func (self *SiteInfo)GetListInit(dir string,D int){
	var err error
	for i:=1;;i++ {
		err= self.GetPageInit(i,dir,D)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (self *SiteInfo)GetPageInit(page int,dir string,Day int) error {
	path:="http://www.sczfcg.com/CmsNewsController.do?"
	uv := url.Values{}
	uv.Add("method","recommendBulletinList")
	uv.Add("rp","25")
	uv.Add("page",fmt.Sprintf("%d",page))
	uv.Add("moreType","provincebuyBulletinMore")
	uv.Add("channelCode","sjcg1")
	path += uv.Encode()
	fmt.Println(path)
	b,err:= self.ClientDO(path)
	if err != nil {
		return err
//		panic(err)
	}
	node,err := goquery.ParseString(string(b))
	if err != nil {
		return err
	//	panic(err)
	}
	li := node.Find(".colsList ul li")
	le :=li.Length()
	fmt.Println(le)
//	isOld:=0
	now := time.Now()
	for i := 0; i < le; i++ {
		v:=li.Eq(i)
		d := &InfoData{v.Find("a").Text(),v.Find("a").Attr("href"),v.Find("span").Text()}

		tis ,err := time.Parse("2006-01-02",d.Time)
		if err != nil {
			panic(err)
		}
		newTime := tis.Add(time.Hour*24*time.Duration(Day))
		if newTime.Before(now) {
			return fmt.Errorf("time out")
		}


		if !self.FindSameOld(d) {
//			isOld ++
		}else{
	//		fmt.Println(d)
			d.SaveData(dir)
			self.NewData = append(self.NewData,d)
			self.OldData = append(self.OldData,d)
		}
	}
//	fmt.Println(isOld,"old")
//	if isOld == le {
//		return fmt.Errorf("isOld %d",isOld)
//	}

	return nil
}
func (self *SiteInfo)GetList(dir string,Page int){
	var err error
	for i:=1;;i++ {
		err= self.GetPage(i,dir,*Day)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
func (self *SiteInfo)GetPage(page int,dir string,Day int) error {

	path:="http://www.sczfcg.com/CmsNewsController.do?"
	uv := url.Values{}
	uv.Add("method","recommendBulletinList")
	uv.Add("rp","25")
	uv.Add("page",fmt.Sprintf("%d",page))
	uv.Add("moreType","provincebuyBulletinMore")
	uv.Add("channelCode","sjcg1")
	path += uv.Encode()
	fmt.Println(path)
	b,err:= self.ClientDO(path)
	if err != nil {
		return err
//		panic(err)
	}
	node,err := goquery.ParseString(string(b))
	if err != nil {
		return err
	//	panic(err)
	}
	li := node.Find(".colsList ul li")
	le :=li.Length()
	fmt.Println(le)
	isOld:=0
	now := time.Now()
	for i := 0; i < le; i++ {
		v:=li.Eq(i)
		d := &InfoData{v.Find("a").Text(),v.Find("a").Attr("href"),v.Find("span").Text()}

		tis ,err := time.Parse("2006-01-02",d.Time)
		if err != nil {
			panic(err)
		}
		newTime := tis.Add(time.Hour*24*time.Duration(Day))
		if newTime.Before(now) {
			return fmt.Errorf("time out")
		}


		if self.FindSameOld(d) {
		//	fmt.Println(d)
			isOld ++
		}else{
	//		fmt.Println(d)
			d.SaveData(dir)
			self.NewData = append(self.NewData,d)
			self.OldData = append(self.OldData,d)
		}
	}
//	fmt.Println(isOld,"old")
	if isOld == le {
		return fmt.Errorf("isOld %d",isOld)
	}

	return nil
}

func main(){
	flag.Parse()
	site := new(SiteInfo)
	err := site.Init(*Dir)
	fmt.Println(err)

	if *Init {
		site.GetListInit(*Dir,*Day)
	}else{
		site.GetList(*Dir,*Day)
	}
	site.SaveNewData(*NewFile)
}
