function extractCover(file, coverCreateCallback) {
    coverCreateCallback = coverCreateCallback || function () { }

    return new Promise((resolve, reject) => {
        const video = document.createElement('video');
        const canvas = document.createElement('canvas');
        const context = canvas.getContext('2d');

        video.preload = 'metadata';

        video.onloadedmetadata = () => {
            canvas.width = video.videoWidth;
            canvas.height = video.videoHeight;

            if (canvas.width == 0 || canvas.height == 0) {
                // 没有正确获取到视频的宽高，设置为默认的值
                canvas.width = 300
                canvas.height = 300
            }

            video.currentTime = 0.5;
        };

        video.onseeked = () => {
            context.drawImage(video, 0, 0, canvas.width, canvas.height);
            canvas.toBlob((blob) => {
                let url = URL.createObjectURL(blob);
                coverCreateCallback(url)
                thumbnailImage(url).then(
                    coverurl => {
                        if (coverurl) {
                            resolve(coverurl)
                        } else {
                            uploadFile(blob, resp => {
                                resolve(resp.url)
                            }, error => {
                                reject(error)
                            }
                            )
                        }
                    }
                ).catch(error => {
                    reject(error)
                })
            }, 'image/webp', 0.8);
        };

        video.onerror = () => {
            reject("读取视频出错")
        }

        video.src = URL.createObjectURL(file);

    })


}


function thumbnailImage(url) {
    return new Promise((resolve, reject) => {
        let img = new Image();
        img.onload = () => {
            const width = img.width;
            const height = img.height;
            const area = width * height;
            const maxArea = 90000;
            if (area > maxArea) {
                const scaleFactor = Math.sqrt(maxArea / area);
                const newWidth = Math.floor(width * scaleFactor);
                const newHeight = Math.floor(height * scaleFactor);

                const canvas = document.createElement('canvas');
                canvas.width = newWidth;
                canvas.height = newHeight;

                const ctx = canvas.getContext('2d');
                ctx.drawImage(img, 0, 0, newWidth, newHeight);
                canvas.toBlob(blob => {
                    uploadFile(blob, resp => {
                        resolve(resp.url)
                    }, error => {
                        reject(error)
                    }
                    )
                }, "image/webp", 0.8)
            } else {
                resolve(undefined)
            }

        }
        img.src = url;
    })

}

function tryCompressImage(file) {
    return new Promise((resolve) => {
        const img = new Image();
        img.onload = () => {
            try {
                const canvas = document.createElement('canvas');
                const ctx = canvas.getContext('2d');

                canvas.width = img.width;
                canvas.height = img.height;
                ctx.drawImage(img, 0, 0);

                canvas.toBlob((blob) => {
                    if (blob) {
                        resolve(blob);
                    } else {
                        resolve(file);
                    }
                }, 'image/webp', 0.8);
            } catch (error) {
                resolve(file);
            }
        };
        img.onerror = () => {
            resolve(file);
        };
        img.src = URL.createObjectURL(file);
    });
}

class Uploader {
    constructor(dom, options) {

        this.dom = dom
        dom.addClass("uploader")
        dom.css("position", "relative")
        dom.css("cursor", "pointer")
        dom.html(`<div style="width: 100%;height: 100%; font-size: 40px;color: #717171;background: #F7F7F7;" class="center"> + </div>`)

        let accept = "*"
        if (options.types) {
            accept = ""
            for (let i = 0; i < options.types.length; i++) {
                let type = options.types[i]
                if (type === "image") {
                    accept = accept + "image/png, image/jpeg,image/png, image/webp, image/gif, "
                }
                if (type === "video") {
                    accept = accept + "video/mp4, "
                }
                if (type === "audio") {
                    accept = accept + "audio/mp3, "
                }
            }

            if (accept !== "") {
                accept = accept.substring(0, accept.length - 2)
            } else {
                accept = "*"
            }
        }

        this.options = options
        let callback = options.callback || function () { }
        let errorcallback = options.errorcallback || function () { }
        let progresscallback = options.progresscallback || function () { }

        let uploadding = false

        let errorcallback0 = error => {
            uploadding = false
            toast("上传失败，点击重试")
            if (progressbar) {
                progressbar.css("background", "#F56C6C")
            }
            if (waitingimg) {
                waitingimg.html(` <img src="/static/images/error.webp" style="width:40px;height:40px;" />`)
            } else {
                waitingimg = $(`<div style="position:absolute;left:0px;right:0px;top:0px;bottom:0px;background: rgba(0, 0, 0, 0.6)" class="center">
                <img src="/static/images/error.webp" style="width:40px;height:40px;" />
            </div>`)
                dom.append(waitingimg)
            }

            waitingimg.click(e => {
                e.stopPropagation()
                this.clear()
            })
            errorcallback(error)
        }

        let progressbar = undefined;
        let waitingimg = undefined;

        dom.click(() => {
            if (uploadding) {
                return
            }
            uploadding = true

            let fileinput = document.createElement("input")
            fileinput.type = "file"
            fileinput.accept = accept
            fileinput.addEventListener('change', async e => {
                const files = e.target.files;
                if (files.length > 0) {
                    let file = files[0]
                    let filetype = file.type
                    let allowUpload = !options.types || options.types.some(x => filetype.indexOf(x) === 0)

                    if (allowUpload) {

                        let isImage = file.type.indexOf("image") === 0
                        let isGif = file.type.indexOf("gif") !== -1
                        let isVideo = file.type.indexOf("video") === 0;
                        let isAudio = file.type.indexOf("audio") === 0;

                        // 添加等待图层
                        waitingimg = $(`<div style="position:absolute;left:0px;right:0px;top:0px;bottom:0px;background: rgba(0, 0, 0, 0.6);" class="center">
                            <img src="/static/images/loading.gif" style="width:40px;height:40px;" />
                        </div>`)

                        if (isImage) {
                            // 添加封面图
                            let coverimg = $(`<img style="width:100%;height:100%;display:block" />`)
                            let imgurl = URL.createObjectURL(file);
                            coverimg.attr("src", imgurl)
                            dom.html("")
                            dom.append(coverimg)
                            dom.append(waitingimg)

                        }

                        let cover = ""
                        let initprogress = 0;

                        if (isVideo) {
                            try {
                                let coverurl = await extractCover(file, url => {
                                    let coverimg = $(`<img style="width:100%;height:100%" />`)
                                    coverimg.attr("src", url)
                                    dom.html("")
                                    dom.append(coverimg)
                                    dom.append(waitingimg)
                                })

                                cover = coverurl
                                initprogress = 10
                                progresscallback(10)

                            } catch (err) {
                                errorcallback0(err)
                                return
                            }

                        }


                        // 当上传的文件大于2M的时候，增加上传进度条
                        if (file.size > 1024 * 1024 * 2) {
                            progressbar = $(`<div class="progressbar"></div>`)
                            dom.append(progressbar)
                        }

                        if (isImage && options.cover && !isGif) {
                            // 需要缩略图
                            try {
                                cover = await thumbnailImage(URL.createObjectURL(file))
                                // 给缩略图的生成以及上传 20的进度
                                initprogress = 10;
                                progresscallback(10)
                            } catch (err) {
                                errorcallback0(err)
                                return
                            }

                        }

                        if (isImage && forceCompressImage && !isGif) {
                            console.log("尝试压缩图片.....")
                            file = await tryCompressImage(file)
                        }

                        // 开始上传了，
                        uploadFile(file, resp => {
                            waitingimg.remove()
                            if (options.cover) {
                                resp.cover = cover || resp.url
                            }
                            this.url = resp

                            this.addDeleteBtn()

                            uploadding = false
                            callback(resp)
                        }, error => {
                            errorcallback0(error)
                        }, progress => {
                            let curprogress = Math.max(initprogress, progress)
                            if (progressbar) {
                                progressbar.css("width", curprogress + "%")
                            }
                            progresscallback(curprogress)
                        })
                    }
                }
            });
            fileinput.addEventListener('cancel', () => {
                uploadding = false
            })
            fileinput.click();
        })
    }

    clear() {
        if (!this.url) {
            this.reset()
        } else {
            this.setInitUrl(this.url)
        }
    }

    reset() {
        this.dom.html(`<div style="width: 100%;height: 100%; font-size: 40px;color: #717171;background: #F7F7F7;" class="center"> + </div>`)
        this.url = undefined
    }

    addDeleteBtn() {
        let options = this.options
        if (options.allowDelete) {
            // 增加删除按钮
            let deleteCallback = options.deleteCallback || function () { }
            let deleteBtn = $(`<div class="deleteicon">
            <?xml version="1.0" standalone="no"?>
                <i class="iconfont icon-delete"></i>
            </div>`)
            this.dom.append(deleteBtn)
            deleteBtn.click(event => {
                event.stopPropagation()
                this.delete()
                deleteCallback()
            })

            // 允许删除状态下，不允许点击重新上传
            this.dom.css("cursor", "auto")
            this.dom.unbind("click")
        }

        if (options.allowSoftDelete) {
            // 增加删除按钮
            let deleteCallback = options.deleteCallback || function () { }
            let deleteBtn = $(`<div class="deleteicon">
             <i class="iconfont icon-delete"></i>
            </div>`)
            this.dom.append(deleteBtn)
            deleteBtn.click(event => {
                event.stopPropagation()
                this.url = undefined
                this.clear()
                deleteCallback()
            })
        }
    }

    setInitUrl(url) {
        this.url = url
        if (url.type === "image") {
            this.dom.html(`<img style="width:100%;height:100%" src="${url.url}" />`)
            this.addDeleteBtn()
        }
        if (url.type === "video") {
            this.dom.html(`<img style="width:100%;height:100%" src="${url.cover}" />`)
            this.addDeleteBtn()
        }
    }

    getUrl() {
        return this.url
    }

    delete() {
        if (this.dom) {
            this.dom.remove()
            this.dom = undefined
        }
    }
}


function calculateMD5(file) {

    var md5 = CryptoJS.algo.MD5.create();

    return new Promise((resolve, reject) => {
        const chunkSize = 1024 * 1024; // 1MB
        const chunks = Math.ceil(file.size / chunkSize);
        let currentChunk = 0;

        const reader = new FileReader();

        reader.onload = function (event) {
            const arrayBuffer = event.target.result;
            const wordArray = CryptoJS.lib.WordArray.create(arrayBuffer);
            md5.update(wordArray)

            currentChunk++;

            if (currentChunk < chunks) {
                loadNextChunk();
            } else {
                resolve(md5.finalize().toString());
            }
        };

        reader.onerror = function (error) {
            reject(error);
        };

        function loadNextChunk() {
            const start = currentChunk * chunkSize;
            const end = Math.min(start + chunkSize, file.size);
            const blob = file.slice(start, end);
            reader.readAsArrayBuffer(blob);
        }

        loadNextChunk();
    });
}


function calculateChunkMD5(file, start, end) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = function (event) {
            const arrayBuffer = event.target.result;
            const wordArray = CryptoJS.lib.WordArray.create(arrayBuffer);
            const md5 = CryptoJS.MD5(wordArray).toString();
            resolve(md5);
        };
        reader.onerror = function (error) {
            reject(error);
        };
        reader.readAsArrayBuffer(file.slice(start, end));
    });
}

function uploadFilePromise(file, uploadUrl, uploadProgresslback) {
    return new Promise((resolve, reject) => {
        let filesize = file.size
        const xhr = new XMLHttpRequest();
        xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable) {
                uploadProgresslback(event.lengthComputable)
            }
        });
        xhr.addEventListener('loadend', () => {
            if (xhr.readyState === 4 && xhr.status === 200) {
                uploadProgresslback(filesize)
                resolve()
            } else {
                reject("上传失败")
            }
        });
        xhr.addEventListener('error', () => {
            reject("上传失败")
        });

        xhr.open('PUT', uploadUrl, true);
        xhr.send(file);
    })
}

function getFileType(file) {
    return file.type.indexOf("image") === 0 ? "image" : (
        file.type.indexOf("video") === 0 ? "video" : (
            file.type.indexOf("audio") === 0 ? "music" : "other"
        )
    )
}

async function uploadFile(file, successcallback, errorcallback, progresscallback) {

    successcallback = successcallback || function () { }
    errorcallback = errorcallback || function () { }
    progresscallback = progresscallback || function () { }

    const timepath = formatDate(new Date(), 'yyyy/MM/dd');
    const newnname = getRandomFileName(file.name, file.type)
    const key = `storage/${timepath}/${newnname}`;
    const getUploadUrl = `/api/v1/file/${key}`;


    fetch(getUploadUrl, {
        method: "POST"
    }).then(resp => {
        resp.json().then(respjson => {
            let uploadUrl = respjson.data
            if (uploadUrl.indexOf("yunpan123_openai") === -1) {

                // 开始上传
                const xhr = new XMLHttpRequest();
                xhr.upload.addEventListener('progress', (event) => {
                    if (event.lengthComputable) {
                        const process = (event.loaded / event.total) * 100;
                        progresscallback(process)
                    }
                });
                xhr.addEventListener('loadend', () => {
                    if (xhr.readyState === 4 && xhr.status === 200) {
                        progresscallback(100)
                        successcallback({
                            type: getFileType(file),
                            url: getUploadUrl
                        })
                    } else {
                        errorcallback("上传失败")
                    }
                });
                xhr.addEventListener('error', () => {
                    errorcallback("上传失败")
                });

                xhr.open('PUT', uploadUrl, true);
                xhr.send(file);
            } else {
                // 123云盘的分片上传
                uploadFileToYunpan123(getUploadUrl, uploadUrl, file, successcallback, errorcallback, progresscallback);
            }

        })
    }).catch(err => {
        console.log(err)
    })

}

async function uploadFileToYunpan123(getUploadUrl, uploadUrl, file, successcallback, errorcallback, progresscallback) {
    let md5 = await calculateMD5(file)
    let preUploadResp = await fetch(`${uploadUrl}&md5=${md5}&size=${file.size}`, {
        method: "PUT"
    })
    const filetype = getFileType(file)
    if (preUploadResp.status === 200) {
        let preUploadRespJson = await preUploadResp.json()
        if (preUploadRespJson.reuse) {
            progresscallback(100)
            successcallback({
                type: filetype,
                url: getUploadUrl
            })
        } else {
            let totalSize = file.size
            // 需要分片上传
            let chunkSize = preUploadRespJson.sliceSize
            let chunkNum = Math.ceil(totalSize / chunkSize)
            let finishedSize = 0

            let chunkMd5s = []
            for (let i = 1; i <= chunkNum; i++) {
                let getUploadUrlResp = await fetch(`${uploadUrl}&sliceNo=${i}&preuploadID=${preUploadRespJson.preuploadID}`)
                if (getUploadUrlResp.status === 200) {
                    let getUploadUrlJson = await getUploadUrlResp.json()
                    let presignedURL = getUploadUrlJson.presignedURL

                    const start = (i - 1) * chunkSize;
                    const end = Math.min(start + chunkSize, file.size);

                    let chunkMd5 = await calculateChunkMD5(file, start, end)
                    chunkMd5s.push(chunkMd5)

                    let chunk = file.slice(start, end);

                    try {
                        await uploadFilePromise(chunk, presignedURL, uploadedSize => {
                            let currentSize = finishedSize + uploadedSize
                            progresscallback((currentSize / totalSize * 100).toFixed(2))
                        })

                        finishedSize = finishedSize + chunk.size;
                        console.log("上传第", i, "个分片成功")
                    } catch (e) {
                        errorcallback("上传切片出错")
                        return
                    }

                } else {
                    errorcallback("获取预上传链接出错")
                    return
                }
            }

            // 分片进行检查
            let checkResp = await fetch(`${uploadUrl}&preuploadID=${preUploadRespJson.preuploadID}`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(chunkMd5s)
            })

            if (checkResp.status === 200) {
                progresscallback(100)
                successcallback({
                    type: filetype,
                    url: getUploadUrl
                })
            } else {
                errorcallback("分片检查失败")
            }

        }
    } else {
        errorcallback("创建文件出错")
    }
}


function getRandomFileName(filename, filetype) {

    var extension = "";

    if (filename) {
        var parts = filename.split('.');
        if (parts.length > 1) {
            extension = parts.pop().toLowerCase();
        }
    } else if (filetype) {
        let arr = filetype.split("/")
        if (arr.length === 2) {
            extension = arr[1]
        }
    } else {
        extension = "webp"
    }

    var randomName = (Math.ceil(Math.random() * 1000) + new Date().getTime() * 1000).toString(36)
    if (extension !== "") {
        randomName += '.' + extension;
    }

    return randomName;
}

function formatDate(inputDate, format) {
    if (!inputDate) return '';

    const padZero = (value) => (value < 10 ? `0${value}` : `${value}`);
    const parts = {
        yyyy: inputDate.getFullYear(),
        MM: padZero(inputDate.getMonth() + 1),
        dd: padZero(inputDate.getDate()),
        HH: padZero(inputDate.getHours()),
        hh: padZero(inputDate.getHours() > 12 ? inputDate.getHours() - 12 : inputDate.getHours()),
        mm: padZero(inputDate.getMinutes()),
        ss: padZero(inputDate.getSeconds()),
        tt: inputDate.getHours() < 12 ? 'AM' : 'PM',
    };

    return format.replace(/yyyy|MM|dd|HH|hh|mm|ss|tt/g, (match) => parts[match]);
}
