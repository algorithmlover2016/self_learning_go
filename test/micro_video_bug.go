package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"unsafe"
	ctrintpb "write_index/protobuf/ctrint_reduce"
	ctrstrpb "write_index/protobuf/ctrstr_reduce"

	"github.com/golang/protobuf/proto"
)

const (
	UINT8_SIZE    = uint32(1)
	UINT32_SIZE   = uint32(4)
	FLOAT32_SIZE  = uint32(4)
	UINT64_SIZE   = uint32(8)
	FLOAT64_SIZE  = uint32(8)
	DOC_ITEM_SIZE = UINT64_SIZE + UINT8_SIZE
	N             = int(unsafe.Sizeof(0))
	TOPIC_ALL_8   = uint64(1111)
	MINIMAL_VIDS  = int(1)
	CTR_VP_PREFIX = string("vu_")
)

type DocItem struct {
	Vid     uint64
	Weight  uint8
	SortVal uint64
}

type MicroVideoItem struct {
	Title       string `json:"title"`
	Vid         string `json:"vid"`
	TitleSign   uint64 `json:"title_sign"`
	Mthid       string `json:"mthid"`
	PlayCnt     uint64 `json:"playcnt"`
	CommentCnt  uint64 `json:"commentcnt"`
	PublishTime uint64 `json:"pubtime"`
}

type TopicItem struct {
	TopicId string   `json:"topicid"`
	Title   string   `json:"title"`
	VidList []string `json:"vidlist"`
}

// Each data structure in the inverted row
type TopicIndexItem struct {
	Title   string
	DocList []*DocItem
}

type ByScoreDescending []*DocItem
type ByScoreAscending []*DocItem

// sort slice by weight
func (items ByScoreDescending) Len() int {
	return len(items)
}
func (items ByScoreDescending) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items ByScoreDescending) Less(i, j int) bool {
	return items[i].SortVal > items[j].SortVal
}

func (items ByScoreAscending) Len() int {
	return len(items)
}

func (items ByScoreAscending) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items ByScoreAscending) Less(i, j int) bool {
	return items[i].SortVal < items[j].SortVal
}

type ByWeightDescending []*DocItem
type ByWeightAscending []*DocItem

// sort slice by weight
func (items ByWeightDescending) Len() int {
	return len(items)
}
func (items ByWeightDescending) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items ByWeightDescending) Less(i, j int) bool {
	return items[i].Weight > items[j].Weight
}

func (items ByWeightAscending) Len() int {
	return len(items)
}

func (items ByWeightAscending) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items ByWeightAscending) Less(i, j int) bool {
	return items[i].Weight < items[j].Weight
}

func choose() binary.ByteOrder {
	x := 0x1234
	p := unsafe.Pointer(&x)
	p2 := (*[N]byte)(p)
	if p2[0] == 0 {
		return binary.BigEndian
	} else {
		return binary.LittleEndian
	}
}

var ByteOrder binary.ByteOrder = choose()
var MicroVideoReshape map[uint64]MicroVideoItem = make(map[uint64]MicroVideoItem, 0)
var TopicHotReshape map[uint64]*TopicIndexItem = make(map[uint64]*TopicIndexItem, 0)
var TopicTimeReshape map[uint64]*TopicIndexItem = make(map[uint64]*TopicIndexItem, 0)
var TopicReshape map[uint64]*TopicIndexItem = make(map[uint64]*TopicIndexItem, 0)
var CtrVoteUpReshape map[string]*ctrstrpb.CtrInfo = make(map[string]*ctrstrpb.CtrInfo, 0)
var CtrIntReshape map[uint64]*ctrintpb.CtrInfo = make(map[uint64]*ctrintpb.CtrInfo, 0)

// read Topic data from file
func LoadTopicData(FileName string,
	MicroVideoReshape map[uint64]MicroVideoItem,
	TopicTimeReshape map[uint64]*TopicIndexItem,
	TopicHotReshape map[uint64]*TopicIndexItem,
	TopicReshape map[uint64]*TopicIndexItem,
	CtrIntReshape map[uint64]*ctrintpb.CtrInfo,
	CtrVoteUpReshape map[string]*ctrstrpb.CtrInfo) {
	fr, err := os.Open(FileName)
	Check(err)
	defer func() {
		if err := fr.Close(); err != nil {
			fmt.Println("close filename %s failed", FileName)
		}
	}()
	scanner := bufio.NewScanner(bufio.NewReader(fr))

	var itemIndex TopicIndexItem
	for scanner.Scan() {
		var itemEle TopicItem
		if err := json.Unmarshal(scanner.Bytes(), &itemEle); err != nil {
			err = fmt.Errorf("Unmarshal Topic data error: %v", err)
			fmt.Fprintln(os.Stderr, err)
		} else {
			var itemIndexForTime TopicIndexItem
			var itemIndexForHot TopicIndexItem

			topicId, err := strconv.ParseUint(itemEle.TopicId, 10, 64)
			if err != nil {
				err = fmt.Errorf("parse Topic topicid from string to uint64 error, topicid is %s, and err is %v", itemEle.TopicId, err)
				fmt.Println(err)
				continue
			}

			itemIndexForTime.Title = itemEle.Title
			itemIndexForHot.Title = itemEle.Title
			var filterRepeatVid map[uint64]bool = make(map[uint64]bool, 0)
			for _, itemStr := range itemEle.VidList {
				item, err := strconv.ParseUint(itemStr, 10, 64)
				if err != nil {
					err = fmt.Errorf("parse Topic VidList.vid from string to uint64 error, vid is %s, and err is %v", itemStr, err)
					fmt.Println(err)
					continue
				}
				if _, ok := filterRepeatVid[item]; ok {
					fmt.Printf("vid %d already exists\n", item)
					continue
				} else {
					filterRepeatVid[item] = true
				}

				var SortTime, SortHot uint64 = 0, 0
				var weightTime, weightHot uint8 = 0, 0
				// search MicroVideoData map[uint64]MicroVideoItem
				videoItem, ok := MicroVideoReshape[item]
				if ok {

					CtrVpVidKey := CTR_VP_PREFIX + itemStr
					if CtrVpVal, ok := CtrVoteUpReshape[CtrVpVidKey]; ok {
						// call the computer score function
						SortTime = videoItem.ComputeScoreForTime(CtrVpVal)
						SortHot = videoItem.ComputeScoreForHot(CtrVpVal)
					} else {
						fmt.Printf("the vid %v doesnot exist in ctr_string\n", item)
					}
					// compute weight
					weightTime = videoItem.ComputeWeightForTime()
					weightHot = videoItem.ComputeWeightForHot()
				} else {
					fmt.Printf("the vid %v doesnot exist in json library\n", item)
					continue
				}
				// storage the vid and weight
				itemIndexForTime.DocList = append(itemIndexForTime.DocList,
					&DocItem{Vid: item, Weight: weightTime, SortVal: SortTime})
				itemIndexForHot.DocList = append(itemIndexForHot.DocList,
					&DocItem{Vid: item, Weight: weightHot, SortVal: SortHot})
			}
			if len(itemIndexForHot.DocList) < MINIMAL_VIDS {
				continue
			}
			// according to Weight to sort DocList slice
			sort.Sort(ByScoreDescending(itemIndexForTime.DocList))
			sort.Sort(ByScoreDescending(itemIndexForHot.DocList))

			weight := uint8(0)
			sortVal := uint64(0)
			itemIndex.DocList = append(itemIndex.DocList, &DocItem{
				Vid: topicId, Weight: weight, SortVal: sortVal})
			TopicTimeReshape[topicId] = &itemIndexForTime
			TopicHotReshape[topicId] = &itemIndexForHot
		}
	}
	TopicReshape[TOPIC_ALL_8] = &itemIndex
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

// compute weight by shift bytes
func (MicroVideoItem *MicroVideoItem) ComputeScoreForTime(CtrVpVal *ctrstrpb.CtrInfo) uint64 {
	var weight uint64 = MicroVideoItem.PublishTime
	return weight
}

// compute weight by shift bytes
func (MicroVideoItem *MicroVideoItem) ComputeWeightForTime() uint8 {
	publishTime := MicroVideoItem.PublishTime
	var weight uint8 = 0
	for index := 7; index >= 0; index-- {
		shiftIndex, err := Int8ToUint8(int8(index))
		if err != nil {
			continue
		}
		shiftBits := shiftIndex * 8
		if tmp_byte := (publishTime >> shiftBits) & 0xff; tmp_byte != 0 {
			weight |= (0x01) << shiftIndex
		}
	}
	return weight
}

// compute weight by shift bytes
func (MicroVideoItem *MicroVideoItem) ComputeScoreForHot(CtrVpVal *ctrstrpb.CtrInfo) uint64 {
	return uint64(*CtrVpVal.Click)
}

// compute weight by shift bytes
func (MicroVideoItem *MicroVideoItem) ComputeWeightForHot() uint8 {
	mthid, err := strconv.ParseUint(MicroVideoItem.Mthid, 10, 64)
	if err != nil {
		err = fmt.Errorf("parse MicroVideoData.mthid from string to uint64 error, vid is %s, and err is %v", MicroVideoItem.Mthid, err)
		fmt.Println(err)
		mthid = 0
	}

	playCnt, commentCnt := MicroVideoItem.PlayCnt, MicroVideoItem.CommentCnt
	var weight uint8 = 0
	for index := 7; index >= 0; index-- {
		shiftIndex, err := Int8ToUint8(int8(index))
		if err != nil {
			continue
		}
		shiftBits := shiftIndex * 8
		if tmp_byte := ((mthid >> shiftBits) | (playCnt >> shiftBits) | (commentCnt >> shiftBits)) & 0xff; tmp_byte != 0 {
			weight |= (0x01) << shiftIndex
		}
	}
	return weight
}

// read MicroVideoDat from file whose format is json
func LoadMicroVideoData(FileName string, MicroVideoReshape map[uint64]MicroVideoItem) {
	fr, err := os.Open(FileName)
	Check(err)
	defer func() {
		if err := fr.Close(); err != nil {
			fmt.Println("close filename %s failed", FileName)
		}
	}()
	scanner := bufio.NewScanner(bufio.NewReader(fr))
	for scanner.Scan() {
		var mvItem MicroVideoItem
		if err := json.Unmarshal(scanner.Bytes(), &mvItem); err != nil {
			err = fmt.Errorf("Unmarshal MicroVideo data error: %v", err)
			fmt.Println(err)
		} else {
			vid, err := strconv.ParseUint(mvItem.Vid, 10, 64)
			if err != nil {
				err = fmt.Errorf("parse vid from string to uint64 error, vid is %s, and err is %v", mvItem.Vid, err)
				fmt.Println(err)
				continue

			}
			MicroVideoReshape[vid] = mvItem
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func Float32ToBytes(num float32) []byte {
	bits := math.Float32bits(num)
	bytes := make([]byte, 4)
	ByteOrder.PutUint32(bytes, bits)
	return bytes
}

func Float64ToBytes(num float64) []byte {
	bits := math.Float64bits(num)
	bytes := make([]byte, 8)
	ByteOrder.PutUint64(bytes, bits)
	return bytes
}

func Uint8ToBytes(num uint8) []byte {
	num_bytes := make([]byte, 1)
	num_bytes[0] = byte(num)
	return num_bytes
}

func Uint32ToBytes(num uint32) []byte {
	num_bytes := make([]byte, 4)
	ByteOrder.PutUint32(num_bytes, num)
	return num_bytes
}

func Uint64ToBytes(num uint64) []byte {
	num_bytes := make([]byte, 8)
	ByteOrder.PutUint64(num_bytes, num)
	return num_bytes
}

func BytesToFloat32(bytes []byte) float32 {
	bits := ByteOrder.Uint32(bytes)
	return math.Float32frombits(bits)
}

func BytesToFloat64(bytes []byte) float64 {
	bits := ByteOrder.Uint64(bytes)
	return math.Float64frombits(bits)
}

func BytesToUint32(bytes []byte) uint32 {
	return ByteOrder.Uint32(bytes)
}

func BytesToUint64(bytes []byte) uint64 {
	return ByteOrder.Uint64(bytes)
}

func Int8ToUint8(num int8) (uint8, error) {
	var unum uint8 = 0
	var err error = nil
	var largeNum int64 = 0
	if num < 0 {
		largeNum -= int64(num)
	} else {
		largeNum = int64(num)
	}

	tmpStr := strconv.FormatInt(largeNum, 10)
	if largeUnum, err := strconv.ParseUint(tmpStr, 10, 64); err == nil {
		unum = uint8(largeUnum)
	} else {
		err = fmt.Errorf("Int8 to Uint8 failed, input is %d, error is %v", num, err)
	}
	return unum, err

}
func Int32ToUint32(num int32, base int) (uint32, error) {
	var unum uint32 = 0
	var err error = nil
	if num >= 0 {
		str := strconv.FormatInt(int64(num), base)
		if unum64, err := strconv.ParseUint(str, base, 32); err == nil {
			unum = uint32(unum64)
		} else {
			fmt.Println("str to uint32 error")
		}
	} else {
		err = fmt.Errorf("input value is negative whose value is:%s", num)
		fmt.Println("input value is smaller than zero")
	}
	return unum, err
}

func Int64ToUint64(num int64, base int) (uint64, error) {
	var unum uint64 = 0
	var err error = nil
	if num >= 0 {
		str := strconv.FormatInt(num, base)
		if unum64, err := strconv.ParseUint(str, base, 64); err == nil {
			unum = unum64
		} else {
			fmt.Println("str to uint64 error")
		}
	} else {
		err = fmt.Errorf("input value is negative whose value is:%s", num)
		fmt.Println("input value is smaller than zero")
	}
	return unum, err
}

func Uint64ToFloat64(num uint64, base int) (float64, error) {
	var fnum float64 = 0
	var err error = nil
	num_str := strconv.FormatUint(num, base)
	if fnum64, err := strconv.ParseFloat(num_str, base); err == nil {
		fnum = fnum64
	} else {
		fmt.Println("str to float64 error")
	}
	return fnum, err
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func WriteIndexDataToFile(buf_fw *bufio.Writer, key []byte, DocItemList []*DocItem) error {
	// get bytes length whose type is uint32
	var err error = nil
	key_len, err := Int32ToUint32(int32(len(key)), 10)
	if err != nil {
		err = fmt.Errorf("change key's length to uint32 failed, key is %s, error is %v", key, err)
		return err
	}
	if len_w, err := buf_fw.Write(Uint32ToBytes(key_len)); err != nil {
		err = fmt.Errorf("key_len write value error, key is %s, error is %v", key, err)
		return err
	} else {
		fmt.Printf("key_len write %d bytes successfully\n", len_w)
	}
	if len_w, err := buf_fw.Write(key); err != nil {
		err = fmt.Errorf("key write value error, key is %s, error is %v", key, err)
		return err
	} else {
		fmt.Printf("key write %d bytes\n", len_w)
	}

	len_tmp, err := Int32ToUint32(int32(len(DocItemList)), 10)
	if err != nil {
		err = fmt.Errorf("change DocItemList's size failed, size is %d, error is%v", len(DocItemList), err)
		return err
	}

	len_list := len_tmp * DOC_ITEM_SIZE
	if len_w, err := buf_fw.Write(Uint32ToBytes(len_list)); err != nil {
		err = fmt.Errorf("list len write value error, size is %d, error is%v", len_list, err)
		return err
	} else {
		fmt.Printf("list len write %d bytes value is %v\n", len_w, len_list)
	}

	for _, DocItemEle := range DocItemList {
		// fmt.Printf("insert DocItem value is: %T\n", ele.Value)

		vid := DocItemEle.Vid
		if len_w, err := buf_fw.Write(Uint64ToBytes(vid)); err != nil {
			fmt.Printf("key write value error %v\n", err)
			continue
		} else {
			fmt.Printf("key write %d bytes\n", len_w)
		}
		weight := DocItemEle.Weight
		if len_w, err := buf_fw.Write(Uint8ToBytes(weight)); err != nil {
			fmt.Printf("weight write value error %v\n", err)
		} else {
			fmt.Printf("weight write %d bytes\n", len_w)
		}
	}
	return err
}

func DumpTopicIndex(FileName string,
	TopicHotReshape map[uint64]*TopicIndexItem,
	TopicTimeReshape map[uint64]*TopicIndexItem,
	TopicReshape map[uint64]*TopicIndexItem) error {
	var err error = nil
	fw, err := os.Create(FileName)
	Check(err)
	defer func() {
		if err = fw.Close(); err != nil {
			err = fmt.Errorf("close Filename %s failed, error is %v", FileName, err)
		}
	}()
	buf_fw := bufio.NewWriter(fw)

	if len_w, err := buf_fw.Write(Uint32ToBytes(DOC_ITEM_SIZE)); err != nil {
		err = fmt.Errorf("DOC_ITEM_SIZE write value error %v\n", err)
		return err
	} else {
		fmt.Printf("DOC_ITEM_SIZE write %d bytes successfully\n", len_w)
	}
	for _, TopicVal := range TopicReshape {
		// first writing key to file, key_len first, and then key_value
		key := []byte("TOPIC_ALL" + "_8")
		if err = WriteIndexDataToFile(buf_fw, key, TopicVal.DocList); err != nil {
			return err
		}
	}

	for TopicHotId, TopicHotVal := range TopicHotReshape {
		// first writing key to file, key_len first, and then key_value
		key := []byte("TOPIC_" + strconv.FormatUint(TopicHotId, 10) + "_HOT_8")
		if err = WriteIndexDataToFile(buf_fw, key, TopicHotVal.DocList); err != nil {
			return err
		}
	}
	buf_fw.Flush()
	for TopicTimeId, TopicTimeVal := range TopicTimeReshape {
		// first writing key to file, key_len first, and then key_value
		key := []byte("TOPIC_" + strconv.FormatUint(TopicTimeId, 10) + "_NEW_8")
		if err = WriteIndexDataToFile(buf_fw, key, TopicTimeVal.DocList); err != nil {
			return err
		}
	}
	buf_fw.Flush()
	return err
}

func LoadCtrIntData(FileName string, CtrIntReshape map[uint64]*ctrintpb.CtrInfo) error {
	var err error = nil
	fr, err := os.Open(FileName)
	Check(err)
	defer func() {
		if err := fr.Close(); err != nil {
			fmt.Println("close filename %s failed", FileName)
		}
	}()
	content, err := ioutil.ReadAll(fr)
	if err != nil {
		err = fmt.Errorf("read file error: %v", err)
		return err
	}
	startIndex := uint64(0)
	dataLen := uint64(len(content))
	var keyLen uint64
	var key uint64
	var valueLen uint64
	var value []byte
	const LEN_SIZE = uint64(UINT64_SIZE)
	for startIndex < dataLen {
		keyLen = BytesToUint64(content[startIndex : startIndex+LEN_SIZE])
		startIndex += LEN_SIZE
		key = BytesToUint64(content[startIndex : startIndex+keyLen])
		startIndex += keyLen
		valueLen = BytesToUint64(content[startIndex : startIndex+LEN_SIZE])
		startIndex += LEN_SIZE
		value = content[startIndex : startIndex+valueLen]
		startIndex += valueLen
		ctrPbPtr := &ctrintpb.CtrInfo{}
		if err := proto.Unmarshal(value, ctrPbPtr); err != nil {
			fmt.Printf("Failed to parse CtrInfo string: %v", err)
		} else {
			// fmt.Printf("key_len:%d, key:%d, value_len:%d, value:%v", keyLen, key, valueLen, *ctrPbPtr)
			CtrIntReshape[key] = ctrPbPtr
		}
	}
	return err
}

func LoadCtrVoteUpData(FileName string, CtrVoteUpReshape map[string]*ctrstrpb.CtrInfo) error {
	var err error = nil
	fr, err := os.Open(FileName)
	Check(err)
	defer func() {
		if err := fr.Close(); err != nil {
			fmt.Println("close filename %s failed", FileName)
		}
	}()
	content, err := ioutil.ReadAll(fr)
	if err != nil {
		err = fmt.Errorf("read file error: %v", err)
		return err
	}
	startIndex := uint64(0)
	dataLen := uint64(len(content))
	var keyLen uint64
	var key []byte
	var valueLen uint64
	const LEN_SIZE = uint64(UINT64_SIZE)
	var value []byte
	for startIndex < dataLen {
		keyLen = BytesToUint64(content[startIndex : startIndex+LEN_SIZE])
		startIndex += LEN_SIZE
		key = content[startIndex : startIndex+keyLen]
		startIndex += keyLen
		valueLen = BytesToUint64(content[startIndex : startIndex+LEN_SIZE])
		startIndex += LEN_SIZE
		value = content[startIndex : startIndex+valueLen]
		startIndex += valueLen
		ctrPbPtr := &ctrstrpb.CtrInfo{}
		if err = proto.Unmarshal(value, ctrPbPtr); err != nil {
			fmt.Printf("Failed to parse CtrInfo string: %v", err)
		} else if strings.Contains(string(key), CTR_VP_PREFIX) {
			CtrVoteUpReshape[string(key)] = ctrPbPtr
		} else {
			fmt.Printf("key: %s is not matched which should contain %s\n",
				string(key), CTR_VP_PREFIX)
		}
	}
	return err
}

func ExecuteProcess(TopicFileName, CtrIntFileName, CtrStrFileName,
	MicroVideoFileName, DumpTopicFileName string) {
	// must load MicroVideoData first to create MicroVideoReshape
	LoadMicroVideoData(MicroVideoFileName, MicroVideoReshape)
	LoadCtrIntData(CtrIntFileName, CtrIntReshape)
	LoadCtrVoteUpData(CtrStrFileName, CtrVoteUpReshape)
	LoadTopicData(TopicFileName, MicroVideoReshape, TopicTimeReshape,
		TopicHotReshape, TopicReshape, CtrIntReshape, CtrVoteUpReshape)
	DumpTopicIndex(DumpTopicFileName, TopicHotReshape, TopicTimeReshape, TopicReshape)
}

func main() {
	var TopicFileName string = "./data/topic_data"
	var MicroVideoFileName string = "./data/content_model_cache.data"
	var DumpTopicFileName string = "./data/dump_topic_index"
	var CtrIntFileName string = "./data/ctr_url_kv"
	var CtrStrFileName string = "./data/vu_vd"
	ExecuteProcess(TopicFileName, CtrIntFileName, CtrStrFileName,
		MicroVideoFileName, DumpTopicFileName)
}
