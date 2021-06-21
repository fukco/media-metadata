## 已支持相机文件格式
* SONY XAVC文件
* 进行中：Canon, FUJI, Panasonic
* 待补充

## 如何使用
### 作为可执行文件
* 输入：1.指定文件 2.指定文件夹
* 输出：1. 控制台输出 2. 达芬奇对应的csv文件
   
### 动态链接库
cgo不支持交叉编译，需要在对应平台进行编译
```shell
go build -buildmode=c-shared -o resolve-metadata.dll .
go build -buildmode=c-shared -o resolve-metadata.so .
```

## support media file format
quicktime(.mov)
mpeg-4(.mp4)

## 解析的box/atom
* 通用 ftyp
* SONY mdat meta -> xml 
* FUJIFILM moov -> udta -> MVTG
* PANASONIC moov -> udta -> PANA

## Davinci Resolve fields
* Camera Type
* Camera Manufacturer
* Camera Serial #
* Camera ID
* Camera Notes
* Camera Format
* Media Type
* Time-lapse Interval
* Camera FPS
* Shutter Type
* Shutter
* ISO
* White Point (Kelvin)
* White Balance Tint
* Camera Firmware
* Lens Type
* Lens Number
* Lens Notes
* Camera Aperture Type
* Camera Aperture
* Focal Point (mm)
* Distance
* Filter
* ND Filter
* Compression Ratio
* Codec Bitrate
* Aspect Ratio Notes
* Gamma Notes
* Color Space Notes

## 其他
1. csv文件导出的方式再导入达芬奇不支持默认的时间码匹配规则，需要选择文件名匹配


