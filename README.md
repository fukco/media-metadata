[![GitHub release](https://img.shields.io/github/release/fukco/media-meta-parser?&style=flat-square)](https://github.com/fukco/media-meta-parser/releases/latest)
[![Paypal Donate](https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square)](https://www.paypal.com/donate/?business=9BGFEVJPEFZAQ&no_recurring=0&currency_code=USD&source=qr)
[![Bilibili](https://img.shields.io/badge/dynamic/json?label=Bilibili&query=%24.data.follower&url=https%3A%2F%2Fapi.bilibili.com%2Fx%2Frelation%2Fstat%3Fvmid%3D26755389&style=social&logo=Bilibili)](https://space.bilibili.com/26755389)
[![Youtube](https://img.shields.io/youtube/channel/subscribers/UCb7NsYnLmtPTn-yddNTcVKA?style=social&label=Youtube)](https://www.youtube.com/channel/UCb7NsYnLmtPTn-yddNTcVKA)

## 已支持相机文件格式
* Atomos
* Canon 
* Fujifilm
* Nikon
* Panasonic
* SONY XAVC文件
* 待补充

## 编译
### 本地编译
#### 编译准备
* 下载源码
* 首次使用，需要安装依赖使用`go mod download`

#### 可执行文件
* 使用`go build -ldflags "-s -w" .`编译生成可执行文件，支持win, mac，windows建议使用GitHub Actions的xgo编译免去安装环境的麻烦，Mac使用xgo编译后会出现不受信任的开发者的问题，需要手动绕过，可以的话本地编译即可
* 使用`./media-meta-parser -h`查看使用帮助

#### 动态链接库
本地编译命令
```shell
go build -buildmode=c-shared -ldflags "-s -w" -o resolve-metadata.dll .
go build -buildmode=c-shared -ldflags "-s -w" -o resolve-metadata.dylib .
```
windows版本依旧建议使用Actions编译

## GitHub Action编译
Fork本仓库，在GitHub的Actions标签页进行相关操作

## 如何使用
### 作为可执行文件
* 首次使用，需要安装依赖使用`go mod download`
* 使用`go build -ldflags "-s -w" .`编译生成可执行文件，支持win, mac，交叉编译需要修GO改环境变量
* 使用`./media-meta-parser -h`查看使用帮助
* 输入参数：1.指定文件 2.指定文件夹
* 输出参数：1. 控制台输出 2. 达芬奇对应的csv文件

## support media file format
quicktime(.mov)
mpeg-4(.mp4)

## DaVinci Resolve fields
| 字段                   | 中文显示      |
|----------------------|-----------|
| Camera Type          | 摄影机类型     |
| Camera Manufacturer  | 摄影机生产厂商   |
| Camera Serial #      | 摄影机序列号    |
| Camera ID            | 摄影机ID     |
| Camera Notes         | 摄影机备注     |
| Camera Format        | 摄影机格式     |
| Media Type           | 媒体类型      |
| Time-lapse Interval  | 延时摄影区间    |
| Camera FPS           | 摄影机帧速率    |
| Shutter Type         | 快门类型      |
| Shutter              | 快门        |
| ISO                  | ISO       |
| White Point (Kelvin) | 白点（开尔文）   |
| White Balance Tint   | 白平衡色调     |
| Camera Firmware      | 摄影机固件     |
| Lens Type            | 摄影机镜头类型   |
| Lens Number          | 摄影机镜头号码   |
| Lens Notes           | 摄影机镜头备注   |
| Camera Aperture Type | 摄影机光圈类型   |
| Camera Aperture      | 摄影机光圈     |
| Focal Point (mm)     | 焦点（毫米）    |
| Distance             | 距离        |
| Filter               | 滤镜        |
| ND Filter            | 中性密度滤镜    |
| Compression Ratio    | 压缩比       |
| Codec Bitrate        | 编解码比特率    |
| Sensor Area Captured | 捕捉到的传感器范围 |
| PAR Notes            | 像素宽高比备注   |
| Aspect Ratio Notes   | 宽高比备注     |
| Gamma Notes          | Gamma备注   |
| Color Space Notes    | 色彩空间备注    |

## 其他
1. csv文件导出的方式再导入达芬奇不支持默认的时间码匹配规则，需要选择文件名匹配
2. console输出存在大量Exif的tag没有name的情况，因为本项目目的是为了提供给达芬奇提取元数据使用，只针对性的做了主要字段的解析，有全部元数据查看需求的可以使用ExifTool等工具,当然如果你觉得哪些字段比较重要需要参照也可以提出来，可以的话我也会加上

