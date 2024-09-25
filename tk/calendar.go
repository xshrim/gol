package tk

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var minYear = 1900
var maxYear = 2049

var dateLayout = "2006-01-02"
var startDateStr = "1900-01-30"

var chineseNumber = []string{"一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "十一", "十二"}
var chineseNumberSpecial = []string{"正", "二", "三", "四", "五", "六", "七", "八", "九", "十", "十一", "腊"}
var monthNumber = map[string]int{"January": 1, "February": 2, "March": 3, "April": 4, "May": 5, "June": 6, "July": 7, "August": 8, "September": 9, "October": 10, "November": 11, "December": 12}

var rulepattern = "^(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n|(?:s\\d+))(\\d+))\\)=\\S+$"
var pattern = "^(solar|lunar)\\((?:m(\\d+)):(ld|(?:d|(?:fw|lw|w(\\d+))n|(?:s\\d+))(\\d+))\\)$"
var monthSolarFestival = map[string][]string{}
var monthLunarFestival = map[string][]string{}

var lunarInfo = []int{
	0x04bd8, 0x04ae0, 0x0a570, 0x054d5, 0x0d260, 0x0d950, 0x16554, 0x056a0, 0x09ad0, 0x055d2,
	0x04ae0, 0x0a5b6, 0x0a4d0, 0x0d250, 0x1d255, 0x0b540, 0x0d6a0, 0x0ada2, 0x095b0, 0x14977,
	0x04970, 0x0a4b0, 0x0b4b5, 0x06a50, 0x06d40, 0x1ab54, 0x02b60, 0x09570, 0x052f2, 0x04970,
	0x06566, 0x0d4a0, 0x0ea50, 0x06e95, 0x05ad0, 0x02b60, 0x186e3, 0x092e0, 0x1c8d7, 0x0c950,
	0x0d4a0, 0x1d8a6, 0x0b550, 0x056a0, 0x1a5b4, 0x025d0, 0x092d0, 0x0d2b2, 0x0a950, 0x0b557,
	0x06ca0, 0x0b550, 0x15355, 0x04da0, 0x0a5d0, 0x14573, 0x052d0, 0x0a9a8, 0x0e950, 0x06aa0,
	0x0aea6, 0x0ab50, 0x04b60, 0x0aae4, 0x0a570, 0x05260, 0x0f263, 0x0d950, 0x05b57, 0x056a0,
	0x096d0, 0x04dd5, 0x04ad0, 0x0a4d0, 0x0d4d4, 0x0d250, 0x0d558, 0x0b540, 0x0b5a0, 0x195a6,
	0x095b0, 0x049b0, 0x0a974, 0x0a4b0, 0x0b27a, 0x06a50, 0x06d40, 0x0af46, 0x0ab60, 0x09570,
	0x04af5, 0x04970, 0x064b0, 0x074a3, 0x0ea50, 0x06b58, 0x055c0, 0x0ab60, 0x096d5, 0x092e0,
	0x0c960, 0x0d954, 0x0d4a0, 0x0da50, 0x07552, 0x056a0, 0x0abb7, 0x025d0, 0x092d0, 0x0cab5,
	0x0a950, 0x0b4a0, 0x0baa4, 0x0ad50, 0x055d9, 0x04ba0, 0x0a5b0, 0x15176, 0x052b0, 0x0a930,
	0x07954, 0x06aa0, 0x0ad50, 0x05b52, 0x04b60, 0x0a6e6, 0x0a4e0, 0x0d260, 0x0ea65, 0x0d530,
	0x05aa0, 0x076a3, 0x096d0, 0x04bd7, 0x04ad0, 0x0a4d0, 0x1d0b6, 0x0d250, 0x0d520, 0x0dd45,
	0x0b5a0, 0x056d0, 0x055b2, 0x049b0, 0x0a577, 0x0a4b0, 0x0aa50, 0x1b255, 0x06d20, 0x0ada0}

var festivalMap = map[string]map[string][]interface{}{
	"solar": {
		"1": {
			"solar(m1:d1)=元旦",
			"solar(m1:d10)=中国110宣传日",
			"solar(m1:lwn1)=世界防治麻风病日",
		},
		"2": {
			"solar(m2:d2)=世界湿地日",
			"solar(m2:d4)=世界抗癌症日",
			"solar(m2:d10)=世界气象日",
			"solar(m2:d14)=情人节",
			"solar(m2:d21)=国际母语日",
			"solar(m2:d29)=国际罕见病日",
		},
		"3": {
			"solar(m3:d3)=全国爱耳日",
			"solar(m3:d8)=国际妇女节",
			"solar(m3:d12)=植树节（中国）",
			"solar(m3:d15)=世界消费者权益日",
			"solar(m3:d21)=世界森林日",
			"solar(m3:d22)=世界水日",
			"solar(m3:d23)=世界气象日",
			"solar(m3:d24)=世界防治结核病日",
		},
		"4": {
			"solar(m4:d1)=愚人节",
			"solar(m4:s345)=寒食节",
			"solar(m4:s456)=清明节",
			"solar(m4:d7)=世界卫生日",
			"solar(m4:d22)=世界地球日",
		},
		"5": {
			"solar(m5:d1)=国际劳动节",
			"solar(m5:d4)=中国青年节",
			"solar(m5:d8)=世界红十字日",
			"solar(m5:d12)=国际护士日",
			"solar(m5:d15)=全国碘缺乏病宣传日",
			"solar(m5:d15)=国际家庭日",
			"solar(m5:d17)=世界高血压日",
			"solar(m5:d17)=世界电信和信息社会日",
			"solar(m5:d18)=国际博物馆日",
			"solar(m5:d19)=中国汶川地震哀悼日",
			"solar(m5:d20)=全国学生营养日",
			"solar(m5:d22)=国际生物多样性日",
			"solar(m5:d31)=世界无烟日",
			"solar(m5:w2n1)=母亲节",
			"solar(m5:w3n1)=全国助残日",
		},
		"6": {
			"solar(m6:d1)=国际儿童节",
			"solar(m6:d5)=世界环境日",
			"solar(m6:d6)=全国爱眼日",
			"solar(m6:d14)=世界献血日",
			"solar(m6:d17)=防治荒漠化和干旱日",
			"solar(m6:d23)=国际奥林匹克日",
			"solar(m6:d25)=全国土地日",
			"solar(m6:d26)=国际禁毒日（反毒品日）",
			"solar(m6:w3n1)=父亲节",
		},
		"7": {
			"solar(m7:d1)=香港回归纪念日",
			"solar(m7:d1)=建党节",
			"solar(m7:d11)=世界人口日",
		},
		"8": {
			"solar(m8:d1)=建军节",
			"solar(m8:d15)=抗日战争纪念日（香港）",
		},
		"9": {
			"solar(m9:d3)=抗日战争胜利纪念日（中国大陆、台湾）",
			"solar(m9:d8)=国际扫盲日",
			"solar(m9:d10)=世界预防自杀日",
			"solar(m9:d10)=教师节",
			"solar(m9:d16)=国际臭氧层保护日",
			"solar(m9:d20)=全国爱牙日（中国大陆）",
			"solar(m9:d21)=世界和平日",
			"solar(m9:d27)=世界旅游日",
			"solar(m9:w4n1)=国际聋人节",
		},
		"10": {
			"solar(m10:d1)=国庆节",
			"solar(m10:d4)=世界动物日",
			"solar(m10:d7)=世界住房日（世界人居日）",
			"solar(m10:d8)=全国高血压日",
			"solar(m10:d8)=世界视觉日",
			"solar(m10:d9)=世界邮政日",
			"solar(m10:d10)=世界精神卫生日",
			"solar(m10:d13)=国际减轻自然灾害日",
			"solar(m10:d15)=国际盲人节",
			"solar(m10:d16)=世界粮食节",
			"solar(m10:d17)=世界消除贫困日",
			"solar(m10:d22)=世界传统医药日",
			"solar(m10:d24)=联合国日",
			"solar(m10:d31)=万圣节",
		},
		"11": {
			"solar(m11:d8)=记者节",
			"solar(m11:d9)=消防宣传日",
			"solar(m11:d14)=世界糖尿病日",
			"solar(m11:d17)=国际大学生节",
			"solar(m11:w4n5)=感恩节",
		},
		"12": {
			"solar(m12:d1)=世界艾滋病日",
			"solar(m12:d3)=世界残疾人日",
			"solar(m12:d9)=世界足球日",
			"solar(m12:d13)=南京大屠杀死难者国家公祭日",
			"solar(m12:d20)=澳门回归纪念日",
			"solar(m12:d21)=国际篮球日",
			"solar(m12:d24)=平安夜",
			"solar(m12:d25)=圣诞节",
			"solar(m12:d26)=毛泽东诞辰",
		},
	},
	"lunar": {
		"1": {
			"lunar(m1:d1)=春节",
			"lunar(m1:d5)=路神生日",
			"lunar(m1:d15)=元宵节",
		},
		"2": {
			"lunar(m2:d2)=龙抬头",
		},
		"5": {
			"lunar(m5:d5)=端午节",
		},
		"6": {
			"lunar(m6:d6)=天贶节",
			"lunar(m6:d6)=姑姑节",
		},
		"7": {
			"lunar(m7:d7)=七夕节",
			"lunar(m7:d15)=中元节(鬼节)",
			"lunar(m7:d30)=地藏节",
		},
		"8": {
			"lunar(m8:d15)=中秋节",
		},
		"9": {
			"lunar(m9:d9)=重阳节",
		},
		"10": {
			"lunar(m10:d1)=祭祖节",
		},
		"12": {
			"lunar(m12:d8)=腊八节",
			"lunar(m12:ld)=除夕",
			"lunar(m12:d23)=北方小年",
			"lunar(m12:d24)=南方小年",
		},
	},
}

// chinaHolidays 中国法定节假日
var chinaHolidays = map[string]map[string]bool{
	"2021": chinaHolidays2021,
	"2022": chinaHolidays2022,
	"2023": chinaHolidays2023,
	"2024": chinaHolidays2024,
}

// chinaHolidays2021 2021年中国法定节假日
var chinaHolidays2021 = map[string]bool{
	// 周五 ～ 周日	元旦
	"2021-01-01": true,
	"2021-01-02": true,
	"2021-01-03": true,
	// 周四 ～ 周三	春节
	"2021-02-11": true,
	"2021-02-12": true,
	"2021-02-13": true,
	"2021-02-14": true,
	"2021-02-15": true,
	"2021-02-16": true,
	"2021-02-17": true,
	// 周六 ～ 周一	清明节
	"2021-04-03": true,
	"2021-04-04": true,
	"2021-04-05": true,
	// 周六 ～ 周三	劳动节
	"2021-05-03": true,
	"2021-05-04": true,
	"2021-05-05": true,
	// 周六 ～ 周一	端午节
	"2021-06-12": true,
	"2021-06-13": true,
	"2021-06-14": true,
	// 周日 ～ 周二	中秋节
	"2021-09-19": true,
	"2021-09-20": true,
	"2021-09-21": true,
	// 周五 ～ 周四	国庆日
	"2021-10-01": true,
	"2021-10-02": true,
	"2021-10-03": true,
	"2021-10-04": true,
	"2021-10-05": true,
	"2021-10-06": true,
	"2021-10-07": true,
}

// chinaHolidays2022 2022年中国法定节假日
var chinaHolidays2022 = map[string]bool{
	// 周六 ～ 周一	元旦
	"2022-01-01": true,
	"2022-01-02": true,
	"2022-01-03": true,
	// 周一 ～ 周日	春节
	"2022-01-31": true,
	"2022-02-01": true,
	"2022-02-02": true,
	"2022-02-03": true,
	"2022-02-04": true,
	"2022-02-05": true,
	"2022-02-06": true,
	// 周日 ～ 周二	清明节
	"2022-04-03": true,
	"2022-04-04": true,
	"2022-04-05": true,
	// 周六 ～ 周三	劳动节
	"2022-04-30": true,
	"2022-05-01": true,
	"2022-05-02": true,
	"2022-05-03": true,
	"2022-05-04": true,
	// 周五 ～ 周日	端午节
	"2022-06-03": true,
	"2022-06-04": true,
	"2022-06-05": true,
	// 周六 ～ 周一	中秋节
	"2022-09-10": true,
	"2022-09-11": true,
	"2022-09-12": true,
	// 周六 ～ 周五	国庆日
	"2022-10-01": true,
	"2022-10-02": true,
	"2022-10-03": true,
	"2022-10-04": true,
	"2022-10-05": true,
	"2022-10-06": true,
	"2022-10-07": true,
	// 元旦
	"2022-12-30": true,
	"2022-12-31": true,
}

// chinaHolidays2023 2023年中国法定节假日
var chinaHolidays2023 = map[string]bool{
	// 元旦
	"2023-01-01": true,
	// 春节
	"2023-01-22": true,
	"2023-01-23": true,
	"2023-01-24": true,
	"2023-01-25": true,
	"2023-01-26": true,
	"2023-01-27": true,
	"2023-01-28": true,
	// 清明节
	"2023-04-03": true,
	"2023-04-04": true,
	"2023-04-05": true,
	// 劳动节
	"2023-04-29": true,
	"2023-04-30": true,
	"2023-05-01": true,
	// 端午节
	"2023-06-22": true,
	"2023-06-23": true,
	"2023-06-24": true,
	// 中秋节
	"2023-09-29": true,
	"2023-09-30": true,
	// 国庆日
	"2023-10-01": true,
	"2023-10-02": true,
	"2023-10-03": true,
	"2023-10-04": true,
	"2023-10-05": true,
	"2023-10-06": true,
}

// chinaHolidays2024 2024年中国法定节假日
// https://www.gov.cn/zhengce/content/202310/content_6911527.htm
var chinaHolidays2024 = map[string]bool{
	// 元旦
	"2024-01-01": true,
	// 春节
	"2024-02-10": true,
	"2024-02-11": true,
	"2024-02-12": true,
	"2024-02-13": true,
	"2024-02-14": true,
	"2024-02-15": true,
	"2024-02-16": true,
	"2024-02-17": true,
	// 清明节
	"2024-04-04": true,
	"2024-04-05": true,
	"2024-04-06": true,
	// 劳动节
	"2024-05-01": true,
	"2024-05-02": true,
	"2024-05-03": true,
	"2024-05-04": true,
	"2024-05-05": true,
	// 端午节
	"2024-06-10": true,
	// 中秋节
	"2024-09-15": true,
	"2024-09-16": true,
	"2024-09-17": true,
	// 国庆日
	"2024-10-01": true,
	"2024-10-02": true,
	"2024-10-03": true,
	"2024-10-04": true,
	"2024-10-05": true,
	"2024-10-06": true,
	"2024-10-07": true,
}

// GetLatestTradingDay 返回最近一个交易日string类型日期：YYYY-mm-dd
func GetLatestTradingDay() string {
	today := time.Now()
	holidays := chinaHolidays[fmt.Sprint(today.Year())]
	day := today
	for {
		date := day.Format("2006-01-02")
		if holidays[date] {
			day = day.AddDate(0, 0, -1)
		} else {
			break
		}
	}
	weekday := day.Weekday()
	switch weekday {
	case time.Saturday:
		day = day.AddDate(0, 0, -1)
	case time.Sunday:
		day = day.AddDate(0, 0, -2)
	}
	return day.Format("2006-01-02")
}

// GetPrevTradingDay 返回前一个交易日string类型日期：YYYY-mm-dd
func GetPrevTradingDay() string {
	today := time.Now()
	holidays := chinaHolidays[fmt.Sprint(today.Year())]
	day := today.AddDate(0, 0, -1)
	for {
		date := day.Format("2006-01-02")
		if holidays[date] {
			day = day.AddDate(0, 0, -1)
		} else {
			break
		}
	}
	weekday := day.Weekday()
	switch weekday {
	case time.Saturday:
		day = day.AddDate(0, 0, -1)
	case time.Sunday:
		day = day.AddDate(0, 0, -2)
	}
	return day.Format("2006-01-02")
}

// GetNextTradingDay 返回下一个交易日string类型日期：YYYY-mm-dd
func GetNextTradingDay() string {
	today := time.Now()
	holidays := chinaHolidays[fmt.Sprint(today.Year())]
	day := today.AddDate(0, 0, 1)
	for {
		date := day.Format("2006-01-02")
		if holidays[date] {
			day = day.AddDate(0, 0, 1)
		} else {
			break
		}
	}
	weekday := day.Weekday()
	switch weekday {
	case time.Saturday:
		day = day.AddDate(0, 0, 2)
	case time.Sunday:
		day = day.AddDate(0, 0, 1)
	}
	return day.Format("2006-01-02")
}

// IsTradingDay 返回当期是否为交易日
func IsTradingDay() bool {
	today := time.Now()
	weekday := today.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	holidays := chinaHolidays[fmt.Sprint(today.Year())]
	if holidays[today.Format("2006-01-02")] {
		return false
	}
	return true
}

// func NewFestival(filename string) *Festival {
// 	if filename == "" {
// 		filename = "./festival.json"
// 	}
// 	getFestivalRule(filename)
// 	return &Festival{filename: filename}
// }

func GetfestivalMap(solarDay string) (festivals []string) {
	festivals = []string{}
	loc, _ := time.LoadLocation("Local")

	getFestivalRule()

	//处理公历节日
	tempDate, _ := time.ParseInLocation(dateLayout, solarDay, loc)
	for _, festival := range processRule(tempDate, monthSolarFestival, false, solarDay) {
		festivals = append(festivals, festival)
	}
	//处理农历节日
	lunarDate, isLeapMonth := SolarToLunar(solarDay)
	if !isLeapMonth {
		tempDate, _ := time.ParseInLocation(dateLayout, lunarDate, loc)
		for _, festival := range processRule(tempDate, monthLunarFestival, true, solarDay) {
			festivals = append(festivals, festival)
		}
	}
	return
}

func getFestivalRule() {
	for key, value := range festivalMap["solar"] {
		for _, item := range value {
			v := item.(string)
			is, err := regexp.MatchString(rulepattern, v)
			if err != nil {
				fmt.Println(err.Error())
			}
			if is {
				if _, ok := monthSolarFestival[key]; ok {
					monthSolarFestival[key] = append(monthSolarFestival[key], v)
				} else {
					temp := []string{v}
					monthSolarFestival[key] = temp
				}
			}
		}
	}

	for key, value := range festivalMap["lunar"] {
		for _, item := range value {
			v := item.(string)
			is, err := regexp.MatchString(rulepattern, v)
			if err != nil {
				fmt.Println(err.Error())
			}
			if is {
				if _, ok := monthLunarFestival[key]; ok {
					monthLunarFestival[key] = append(monthLunarFestival[key], v)
				} else {
					temp := []string{v}
					monthLunarFestival[key] = temp
				}
			}
		}
	}
}

func processRule(date time.Time, ruleMap map[string][]string, isLunar bool, solarDay string) []string {
	festivals := []string{}
	year := int(date.Year())
	month := strconv.Itoa(int(date.Month()))
	day := strconv.Itoa(date.Day())
	rules := ruleMap[month]
	for _, rule := range rules {
		items := strings.Split(rule, "=")
		reg, _ := regexp.Compile(pattern)
		subMatch := reg.FindStringSubmatch(items[0])
		festivalMonth := subMatch[2]
		if strings.HasPrefix(subMatch[3], "s456") && !isLunar { //特殊处理清明节
			festivalDay := getQingMingFestival(year)
			if month == festivalMonth && day == festivalDay {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "s345") && !isLunar { //特殊处理寒食节，为清明节前一天
			festivalDay := getQingMingFestival(year)
			intValue, err := strconv.Atoi(festivalDay)
			if err != nil {
				fmt.Print(err.Error())
				continue
			}
			festivalDay = strconv.Itoa(intValue - 1)
			if month == festivalMonth && day == festivalDay {
				festivals = append(festivals, items[1])
			}
		} else if strings.HasPrefix(subMatch[3], "d") {
			festivalDay := subMatch[5]
			if month == festivalMonth && day == festivalDay {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "w") {
			festivalWeek, _ := strconv.Atoi(subMatch[3][1:2])
			festivalDayOfWeek, _ := strconv.Atoi(subMatch[3][3:4])
			week := 0
			tempDayOfWeek := getDayOfWeekOnFirstDayOfMonth(date)
			//特殊处理感恩节，感恩节（m11:w4n5）的计算，不是第4周周4，而是第4个周四，如果第一个周没有周四，就不算第一周
			if compareWeek(tempDayOfWeek, festivalDayOfWeek) && strings.HasPrefix(subMatch[3], "w4n5") {
				week = weekOfMonth(date) - 1
			} else {
				week = weekOfMonth(date)
			}
			dayOfWeek := (int(date.Weekday()) + 1) % 7
			if festivalWeek == week && festivalDayOfWeek == dayOfWeek {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "lw") {
			festivalDayOfWeek, _ := strconv.Atoi(subMatch[3][3:4])
			if isDayOfLastWeeekInTheMonth(date, festivalDayOfWeek) {
				festivals = append(festivals, items[1])
			}
			continue
		} else if strings.HasPrefix(subMatch[3], "ld") && isLunar { //特殊处理除夕节日
			if month == "12" && day == "29" {
				nextLunarDay := lunarDateAddOneDay(solarDay)
				newMonth := strconv.Itoa(int(nextLunarDay.Month()))
				if month != newMonth {
					festivals = append(festivals, items[1])
				}
			} else if month == "12" && day == "30" {
				festivals = append(festivals, items[1])
			}
			continue
		}
	}
	return festivals
}

// 清明节算法 公式：int((yy*d+c)-(yy/4.0)) 公式解读：y=年数后2位，d=0.2422，1=闰年数，21世纪c=4081，20世纪c=5.59
func getQingMingFestival(year int) string {
	var val float64
	if year >= 2000 { //21世纪
		val = 4.81
	} else { //20世纪
		val = 5.59
	}
	d := float64(year % 100)
	day := int(d*0.2422 + val - float64(int(d)/4))
	return strconv.Itoa(day)
}

func lunarDateAddOneDay(solarDay string) time.Time {
	tempDate, err := time.Parse(dateLayout, solarDay)
	if err != nil {
		fmt.Println(err.Error())
	}
	dayDuaration, _ := time.ParseDuration("24h")
	nextDate := tempDate.Add(dayDuaration)
	lunarDate, _ := SolarToLunar(nextDate.Format(dateLayout))
	nexLunarDay, err := time.Parse(dateLayout, lunarDate)
	if err != nil {
		fmt.Println(err.Error())
	}
	return nexLunarDay
}

func weekOfMonth(now time.Time) int {
	beginningOfTheMonth := time.Date(now.Year(), now.Month(), 1, 1, 1, 1, 1, time.UTC)
	_, thisWeek := now.ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()
	return 1 + thisWeek - beginningWeek
}

func isLeapYear(year int) bool {
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		return true
	}
	return false
}

func isDayOfLastWeeekInTheMonth(now time.Time, weekNumber int) bool {
	var endDayOfMonth time.Time
	year := now.Year()
	month := int(now.Month())
	isLeap := isLeapYear(year)
	if month == 2 {
		if isLeap {
			endDayOfMonth = time.Date(now.Year(), now.Month(), 29, 23, 59, 59, 1, time.UTC)
		} else {
			endDayOfMonth = time.Date(now.Year(), now.Month(), 28, 23, 59, 59, 1, time.UTC)
		}
	} else if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		endDayOfMonth = time.Date(now.Year(), now.Month(), 31, 23, 59, 59, 1, time.UTC)
	} else {
		endDayOfMonth = time.Date(now.Year(), now.Month(), 30, 23, 59, 59, 1, time.UTC)
	}
	_, lastWeekOfMonth := endDayOfMonth.ISOWeek()
	_, nowWeekOfMonth := now.ISOWeek()
	dayOfWeek := (int(endDayOfMonth.Weekday()) + 1) % 7
	if dayOfWeek > weekNumber && lastWeekOfMonth > nowWeekOfMonth {
		dayDuaration, _ := time.ParseDuration("-24h")
		endDayOfMonth = endDayOfMonth.Add(dayDuaration * time.Duration(7))
		_, lastWeekOfMonth = endDayOfMonth.ISOWeek()
	}
	if lastWeekOfMonth == nowWeekOfMonth {
		nowDayOfWeek := (int(now.Weekday()) + 1) % 7
		if nowDayOfWeek == weekNumber {
			return true
		}
	}
	return false
}

func compareWeek(first int, second int) bool {
	if first-1 == 0 {
		first = 7
	} else {
		first = first - 1
	}
	if second-1 == 0 {
		second = 7
	} else {
		second = second - 1
	}

	if first >= second {
		return true
	} else {
		return false
	}
}

// 星期日：1 星期一：2 类推
func getDayOfWeekOnFirstDayOfMonth(date time.Time) int {
	date = getFirstDateOfMonth(date)
	dayOfWeek := (int(date.Weekday()) + 1) % 7
	return dayOfWeek
}

func getFirstDateOfMonth(d time.Time) time.Time {
	tempDate := time.Date(d.Year(), d.Month(), d.Day(), d.Hour(), d.Minute(), d.Second(), d.Nanosecond(), d.Location())
	d = tempDate.AddDate(0, 0, -d.Day()+1)
	return getZeroTime(d)
}

func getLastDateOfMonth(d time.Time) time.Time {
	return getFirstDateOfMonth(d).AddDate(0, 1, -1)
}

func getZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func LunarToSolar(date string, leapMonthFlag bool) string {
	date, offset := dealWithSpecialFebruaryDate(date)
	loc, _ := time.LoadLocation("Local")
	lunarTime, err := time.ParseInLocation(dateLayout, date, loc)
	if err != nil {
		fmt.Println(err.Error())
	}
	lunarYear := lunarTime.Year()
	lunarMonth := monthNumber[lunarTime.Month().String()]
	lunarDay := lunarTime.Day()
	err = checkLunarDate(lunarYear, lunarMonth, lunarDay, leapMonthFlag)

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	for i := minYear; i < lunarYear; i++ {
		yearDaysCount := getYearDays(i) // 求阴历某年天数
		offset += yearDaysCount
	}
	//计算该年闰几月
	leapMonth := getLeapMonth(lunarYear)
	if leapMonthFlag && leapMonth != lunarMonth {
		panic("您输入的闰月标志有误！")
	}
	if leapMonth == 0 || (lunarMonth < leapMonth) || (lunarMonth == leapMonth && !leapMonthFlag) {
		for i := 1; i < lunarMonth; i++ {
			tempMonthDaysCount := getMonthDays(lunarYear, uint(i))
			offset += tempMonthDaysCount
		}

		// 检查日期是否大于最大天
		if lunarDay > getMonthDays(lunarYear, uint(lunarMonth)) {
			panic("不合法的农历日期！")
		}
		offset += lunarDay // 加上当月的天数
	} else { //当年有闰月，且月份晚于或等于闰月
		for i := 1; i < lunarMonth; i++ {
			tempMonthDaysCount := getMonthDays(lunarYear, uint(i))
			offset += tempMonthDaysCount
		}
		if lunarMonth > leapMonth {
			temp := getLeapMonthDays(lunarYear) // 计算闰月天数
			offset += temp                      // 加上闰月天数

			if lunarDay > getMonthDays(lunarYear, uint(lunarMonth)) {
				panic("不合法的农历日期！")
			}
			offset += lunarDay
		} else { // 如果需要计算的是闰月，则应首先加上与闰月对应的普通月的天数
			// 计算月为闰月
			temp := getMonthDays(lunarYear, uint(lunarMonth)) // 计算非闰月天数
			offset += temp

			if lunarDay > getLeapMonthDays(lunarYear) {
				panic("不合法的农历日期！")
			}
			offset += lunarDay
		}
	}

	myDate, err := time.ParseInLocation(dateLayout, startDateStr, loc)
	if err != nil {
		fmt.Println(err.Error())
	}

	myDate = myDate.AddDate(0, 0, offset)
	return myDate.Format(dateLayout)
}

func dealWithSpecialFebruaryDate(date string) (string, int) {
	items := strings.Split(date, "-")
	year, _ := strconv.Atoi(items[0])
	if items[1] == "02" {
		if (year/4 == 0 && year/100 != 0) || (year/400 == 0) {
			if items[2] == "30" {
				return items[0] + "-" + items[1] + "-29", 1
			}
		} else {
			if items[2] == "30" {
				return items[0] + "-" + items[1] + "-28", 2
			}
			if items[2] == "29" {
				return items[0] + "-" + items[1] + "-28", 1
			}
		}
	}
	return date, 0
}

func SolarToChineseLunar(date string) string {
	lunarYear, lunarMonth, lunarDay, leapMonth, leapMonthFlag := calculateLunar(date)
	result := cyclical(lunarYear) + "年"
	if leapMonthFlag && (lunarMonth == leapMonth) {
		result += "闰"
	}
	result += chineseNumberSpecial[lunarMonth-1] + "月"
	result += chineseDayString(lunarDay) + "日"
	return result
}

func SolarToSimpleLunar(date string) string {
	lunarYear, lunarMonth, lunarDay, leapMonth, leapMonthFlag := calculateLunar(date)
	result := strconv.Itoa(lunarYear) + "年"
	if leapMonthFlag && (lunarMonth == leapMonth) {
		result += "闰"
	}
	if lunarMonth < 10 {
		result += "0" + strconv.Itoa(lunarMonth) + "月"
	} else {
		result += strconv.Itoa(lunarMonth) + "月"
	}
	if lunarDay < 10 {
		result += "0" + strconv.Itoa(lunarDay) + "日"
	} else {
		result += strconv.Itoa(lunarDay) + "日"
	}
	return result
}

func SolarToLunar(date string) (string, bool) {
	lunarYear, lunarMonth, lunarDay, leapMonth, leapMonthFlag := calculateLunar(date)
	result := strconv.Itoa(lunarYear) + "-"
	if lunarMonth < 10 {
		result += "0" + strconv.Itoa(lunarMonth) + "-"
	} else {
		result += strconv.Itoa(lunarMonth) + "-"
	}
	if lunarDay < 10 {
		result += "0" + strconv.Itoa(lunarDay)
	} else {
		result += strconv.Itoa(lunarDay)
	}

	if leapMonthFlag && (lunarMonth == leapMonth) {
		return result, true
	} else {
		return result, false
	}
}

func calculateLunar(date string) (lunarYear, lunarMonth, lunarDay, leapMonth int, leapMonthFlag bool) {
	loc, _ := time.LoadLocation("Local")
	i := 0
	temp := 0
	leapMonthFlag = false
	isLeapYear := false

	myDate, err := time.ParseInLocation(dateLayout, date, loc)
	if err != nil {
		fmt.Println(err.Error())
	}
	startDate, err := time.ParseInLocation(dateLayout, startDateStr, loc)
	if err != nil {
		fmt.Println(err.Error())
	}

	offset := daysBwteen(myDate, startDate)
	for i = minYear; i < maxYear; i++ {
		temp = getYearDays(i) //求当年农历年天数
		if offset-temp < 1 {
			break
		} else {
			offset -= temp
		}
	}
	lunarYear = i

	leapMonth = getLeapMonth(lunarYear) //计算该年闰哪个月

	//设定当年是否有闰月
	if leapMonth > 0 {
		isLeapYear = true
	} else {
		isLeapYear = false
	}

	for i = 1; i <= 12; i++ {
		if i == leapMonth+1 && isLeapYear {
			temp = getLeapMonthDays(lunarYear)
			isLeapYear = false
			leapMonthFlag = true
			i--
		} else {
			temp = getMonthDays(lunarYear, uint(i))
		}
		offset -= temp
		if offset <= 0 {
			break
		}
	}
	offset += temp
	lunarMonth = i
	lunarDay = offset
	return
}

func checkLunarDate(lunarYear, lunarMonth, lunarDay int, leapMonthFlag bool) error {
	if (lunarYear < minYear) || (lunarYear > maxYear) {
		return fmt.Errorf("非法农历年份！")
	}
	if (lunarMonth < 1) || (lunarMonth > 12) {
		return fmt.Errorf("非法农历月份！")
	}
	if (lunarDay < 1) || (lunarDay > 30) { // 中国的月最多30天
		return fmt.Errorf("非法农历天数！")
	}

	leap := getLeapMonth(lunarYear) // 计算该年应该闰哪个月
	if (leapMonthFlag == true) && (lunarMonth != leap) {
		return fmt.Errorf("非法闰月！")
	}
	return nil
}

// 计算该月总天数
func getMonthDays(lunarYeay int, month uint) int {
	if (month > 31) || (month < 0) {
		fmt.Println("error month")
	}
	// 0X0FFFF[0000 {1111 1111 1111} 1111]中间12位代表12个月，1为大月，0为小月
	bit := 1 << (16 - month)
	if ((lunarInfo[lunarYeay-1900] & 0x0FFFF) & bit) == 0 {
		return 29
	} else {
		return 30
	}
}

// 计算阴历年的总天数
func getYearDays(year int) int {
	sum := 29 * 12
	for i := 0x8000; i >= 0x8; i >>= 1 {
		if (lunarInfo[year-1900] & 0xfff0 & i) != 0 {
			sum++
		}
	}
	return sum + getLeapMonthDays(year)
}

// 计算阴历年闰月多少天
func getLeapMonthDays(year int) int {
	if getLeapMonth(year) != 0 {
		if (lunarInfo[year-1900] & 0xf0000) == 0 {
			return 29
		} else {
			return 30
		}
	} else {
		return 0
	}
}

// 计算阴历年闰哪个月 1-12 , 没闰传回 0
func getLeapMonth(year int) int {
	return (int)(lunarInfo[year-1900] & 0xf)
}

// 计算差的天数
func daysBwteen(myDate time.Time, startDate time.Time) int {
	subValue := float64(myDate.Unix()-startDate.Unix())/86400.0 + 0.5
	return int(subValue)
}

func cyclicalm(num int) string {
	tianGan := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	diZhi := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	animals := []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	return tianGan[num%10] + diZhi[num%12] + animals[num%12]
}

func cyclical(year int) string {
	num := year - 1900 + 36
	return cyclicalm(num)
}

func chineseDayString(day int) string {
	chineseTen := []string{"初", "十", "廿", "三"}
	n := 0
	if day%10 == 0 {
		n = 9
	} else {
		n = day%10 - 1
	}
	if day > 30 {
		return ""
	}
	if day == 20 {
		return "二十"
	} else if day == 10 {
		return "初十"
	} else {
		return chineseTen[day/10] + chineseNumber[n]
	}
}

func Week(datetime string) (y, w int) {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", datetime, time.Local)
	if err != nil {
		return 0, 0
	}
	//获取这个时间的基于这一年有多少天了
	yearDay := t.YearDay()
	//获取上一年的最后一天
	yesterdayYearEndDay := t.AddDate(0, 0, -yearDay)
	//获取上一年最后一天是星期几
	dayInWeek := int(yesterdayYearEndDay.Weekday())
	//第一周的总天数,默认是7天
	firstWeekDays := 7
	//如果上一年最后一天不是星期天，则第一周总天数是7-dayInWeek
	if dayInWeek != 0 {
		firstWeekDays = 7 - dayInWeek
	}
	week := 0
	//如果这一年的总天数小于第一周总天数，则是第一周，否则按照这一年多少天减去第一周的天数除以7+1 但是要考虑这一天减去第一周天数除以7会取整型，
	//所以需要处理两个数取余之后是否大于0，如果大于0 则多加一天，这样自然周就算出来了。
	if yearDay <= firstWeekDays {
		week = 1
	} else {
		plusDay := 0
		if (yearDay-firstWeekDays)%7 > 0 {
			plusDay = 1
		}
		week = (yearDay-firstWeekDays)/7 + 1 + plusDay
	}
	return t.Year(), week
}
