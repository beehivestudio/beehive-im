package comm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/* 字段序号 */
const (
	IP_DICT_START_IP_IDX = 0  // 开始IP
	IP_DICT_END_IP_IDX   = 1  // 结束IP
	IP_DICT_COUNTRY_IDX  = 2  // 国家
	IP_DICT_OPID_IDX     = 6  // 运营商ID
	IP_DICT_NAME_IDX     = -1 // 运营商名称
)

/* IP字典 */
type IpDict struct {
	items []IpDictItem // IP列表
	num   int          // IP列表长度
}

/* IP项 */
type IpDictItem struct {
	start_ip uint32 // 起始IP
	end_ip   uint32 // 结束IP
	opid     uint32 // 运营商ID
	operator string // 运营商名称
	nation   string // 国家或地区名
}

/******************************************************************************
 **函数名称: Ipv4Str2Uint32
 **功    能: 将IPv4字串转换成整型数字
 **输入参数:
 **     ipv4: IPv4字串
 **输出参数: NONE
 **返    回: IP字典
 **实现描述: 将格式为"36.122.133.235"的字串转换成数字
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.25 07:53:48 #
 ******************************************************************************/
func Ipv4Str2Uint32(ipv4 string) uint32 {
	segment := strings.Split(ipv4, ".")
	num := len(segment)
	if 4 != num {
		return 0
	}

	ip_int := uint32(0)
	for idx := 0; idx < num; idx++ {
		digit, _ := strconv.ParseInt(segment[idx], 0, 32)
		if digit > 255 {
			return 0
		}
		d := digit << uint32(8*(num-idx-1))
		ip_int += uint32(d)
	}
	return ip_int
}

/******************************************************************************
 **函数名称: LoadIpDict
 **功    能: 加载IP字典
 **输入参数:
 **     path: 字典路径
 **输出参数: NONE
 **返    回: IP字典
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.25 07:37:49 #
 ******************************************************************************/
func LoadIpDict(path string) (dict *IpDict, err error) {
	fp, err := os.Open(path)
	if nil != err {
		return nil, err
	}

	defer fp.Close()

	dict = &IpDict{}

	dict.num = 0
	dict.items = make([]IpDictItem, 0)

	idx := 0
	rd := bufio.NewReader(fp)
	for {
		line, _, err := rd.ReadLine()
		if io.EOF == err {
			break
		} else if nil != err {
			return nil, err
		}

		idx += 1
		/* > 解析有效数据 */
		segment := strings.Split(string(line), ",")
		if len(segment) < 5 {
			errmsg := fmt.Sprintf("Data isn't right! line:%d", idx)
			return nil, errors.New(errmsg)
		}

		var item IpDictItem

		item.start_ip = Ipv4Str2Uint32(segment[IP_DICT_START_IP_IDX]) /* 起始IP */
		item.end_ip = Ipv4Str2Uint32(segment[IP_DICT_END_IP_IDX])     /* 结束IP */
		item.opid = Ipv4Str2Uint32(segment[IP_DICT_OPID_IDX])         /* 运营商ID */
		item.operator = segment[len(segment)-1]                       /* 运营商名称 */
		item.nation = segment[IP_DICT_COUNTRY_IDX]                    /* 国家或地区 */
		if 0 == item.start_ip || 0 == item.end_ip ||
			"" == item.nation || "" == item.operator {
			errmsg := fmt.Sprintf("Data isn't right! line:%d", idx)
			return nil, errors.New(errmsg)
		} else if idx > 1 {
			//fmt.Printf("idx:%d len:%d ip:%d\n", idx, len(dict.items), dict.items[idx-2].end_ip)
			if item.start_ip <= dict.items[idx-2].end_ip { /* 检测IP是否存在乱序的情况 */
				errmsg := fmt.Sprintf("Ip addr less than last line! line:%d", idx)
				return nil, errors.New(errmsg)
			}
		}

		dict.items = append(dict.items, item)
	}

	dict.num = idx

	return dict, nil
}

/******************************************************************************
 **函数名称: Query
 **功    能: 查询IP字典
 **输入参数:
 **     ipv4: IP地址
 **输出参数: NONE
 **返    回: IP选项(国家+运营商)
 **实现描述: 二分查找算法
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.25 23:18:18 #
 ******************************************************************************/
func (dict *IpDict) Query(ipv4 string) *IpDictItem {
	var low, mid, high int

	if 0 == dict.num {
		return nil
	}

	ip_int := Ipv4Str2Uint32(ipv4)
	if 0 == ip_int {
		return nil
	}

	low = 0
	high = dict.num - 1
	for low <= high {
		mid = (low + high) / 2
		if (dict.items[mid].start_ip <= ip_int) && (dict.items[mid].end_ip >= ip_int) {
			return &dict.items[mid] // found
		}

		for ip_int < dict.items[mid].start_ip && low <= high {
			high = mid - 1
			mid = (low + high) / 2
		}

		for dict.items[mid].end_ip < ip_int && low <= high {
			low = mid + 1
			mid = (low + high) / 2
		}
	}

	return nil
}

/* 获取国家或地区 */
func (item *IpDictItem) GetNation() string {
	return item.nation
}

/* 获取运营商ID */
func (item *IpDictItem) GetOpid() uint32 {
	return item.opid
}

/* 获取运营商名 */
func (item *IpDictItem) GetOperator() string {
	return item.operator
}
