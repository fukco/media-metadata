## 已支持相机文件格式
* Atomos
* Canon 
* Fujifilm
* Panasonic
* SONY XAVC文件
* 待补充

## 如何使用
### 作为可执行文件
* 使用`go build .`编译生成可执行文件，支持win, mac，交叉编译需要修GO改环境变量
* 使用`./media-meta-parser -h`查看使用帮助
* 输入参数：1.指定文件 2.指定文件夹
* 输出参数：1. 控制台输出 2. 达芬奇对应的csv文件
   
### 动态链接库
cgo不支持交叉编译，需要在对应平台进行编译
```shell
go build -buildmode=c-shared -o resolve-metadata.dll .
go build -buildmode=c-shared -o resolve-metadata.so .
```

## support media file format
quicktime(.mov)
mpeg-4(.mp4)

## DaVinci Resolve fields
|  字段   | 中文显示  |
|  ----  | ----  |
| Camera Type | 摄影机类型 |
| Camera Manufacturer | 摄影机生产厂商 |
| Camera Serial # | 摄影机序列号 |
| Camera ID | 摄影机ID |
| Camera Notes | 摄影机备注 |
| Camera Format | 摄影机格式 |
| Media Type | 媒体类型 |
| Time-lapse Interval | 延时摄影区间 |
| Camera FPS | 摄影机帧速率 |
| Shutter Type | 快门类型 |
| Shutter | 快门 |
| ISO | ISO |
| White Point (Kelvin) | 白点（开尔文） |
| White Balance Tint | 白平衡色调 |
| Camera Firmware | 摄影机固件 |
| Lens Type | 摄影机镜头类型 |
| Lens Number | 摄影机镜头号码 |
| Lens Notes | 摄影机镜头备注 |
| Camera Aperture Type | 摄影机光圈类型 |
| Camera Aperture | 摄影机光圈 |
| Focal Point (mm) | 焦点（毫米） |
| Distance | 距离 |
| Filter | 滤镜 |
| ND Filter | 中性密度滤镜 |
| Compression Ratio | 压缩比 |
| Codec Bitrate | 编解码比特率 |
| Aspect Ratio Notes | 宽高比备注 |
| Gamma Notes | Gamma备注 |
| Color Space Notes | 色彩空间备注 |

## 其他
1. csv文件导出的方式再导入达芬奇不支持默认的时间码匹配规则，需要选择文件名匹配
2. console输出存在大量Exif的tag没有name的情况，因为本项目目的是为了提供给达芬奇提取元数据使用，只针对性的做了主要字段的解析，有全部元数据查看需求的可以使用ExifTool等工具,当然如果你觉得哪些字段比较重要需要参照也可以提出来，可以的话我也会加上

