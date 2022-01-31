const client = new WebTorrent();

const videoExt = ["mp4", "ogg", "webm"];

const audioExt = ["mp3", "wav", "flac", "m3u"];

const imageExt = ["png", "jpg", "jpeg", "webp"];

const serverURL = ["https://wt-server.dnlab.net/1/", "https://wt-server.dnlab.net/2/"];

const trackerURL = ["wss://tracker.lili.ac", "wss://tracker.btorrent.xyz", "wss://tracker.openwebtorrent.com"];

$("#uri").keydown(function (e) {
    if (e.keyCode == "13") {
        e.preventDefault();
        let uri = $(this).val();
        sendURI(uri);
    }
});

$("#upload-select").click(function (e) {
    $("#upload-torrent").click();
});

$("#upload-torrent").change(function (e) {
    $(".mdui-dialog-content").text($("#upload-torrent")[0].files[0].name);
});

$("#upload-submit").click(function (e) {
    if ($("#upload-torrent")[0].files[0] == undefined) {
        log("还没选择要上传的文件!");
        return;
    }
    uploadTorrent();
});

$("#share-button").click(function (e) {
    urlsuffix = "/?";
    client.torrents.forEach(function (t) {
        urlsuffix += `uri=${uriEncode(t.magnetURI)}&`;
    });
    urlsuffix = urlsuffix.substr(0, urlsuffix.length - 1);
    res = window.location.origin + urlsuffix;
    $("#share-url").val(res);
});

$("#share-url").click(function (e) {
    $("#share-url").select();
    document.execCommand("copy");
    mdui.snackbar({
        message: "已复制到剪贴板",
        position: "right-bottom",
    });
});

initPage();

function uriEncode(strin) {
    return strin.replaceAll("?", "%3F").replaceAll("=", "%3D").replaceAll("&", "%26");
}

function uriDecode(strin) {
    return strin.replaceAll("%26", "&").replaceAll("%3D", "=").replaceAll("%3F", "?");
}

// 检测url并加载链接
function initPage() {
    let mList = getQueryVariable("uri");
    if (mList.length == 0) {
        return;
    }
    mList.forEach(function (strin) {
        sendURI(uriDecode(strin));
    });
}

function getQueryVariable(variable) {
    let res = [];
    let query = window.location.search.substring(1);
    let vars = query.split("&");
    for (let i = 0; i < vars.length; i++) {
        let pair = vars[i].split("=");
        if (pair[0] == variable) {
            res.push(pair[1]);
        }
    }
    return res;
}

function getExt(filename) {
    return filename.substr(filename.lastIndexOf(".") + 1);
}

function getName(filename) {
    return filename.substring(0, filename.lastIndexOf("."));
}

function isExt(filename, fileExt) {
    let ext = filename.substr(filename.lastIndexOf(".") + 1);
    return fileExt.indexOf(ext.toLowerCase()) != -1;
}

function uploadTorrent() {
    $("#introduce-area").addClass("mdui-hidden");
    log("开始上传文件");
    let okflag = 0;
    let formData = new FormData();
    formData.append("torrent", $("#upload-torrent")[0].files[0]);
    for (let i = 0; i < serverURL.length; i++) {
        $.ajax({
            url: serverURL[i] + "api/add/torrent",
            type: "post",
            data: formData,
            contentType: false,
            processData: false,
            success: function (result) {
                if (result.response == 200) {
                    if (okflag == 0) {
                        okflag = 1;
                        log("torrent 上传成功");
                        startMagnet(result.magnet);
                    }
                } else {
                    log("torrent 上传失败: " + serverURL[i]);
                }
            },
        });
    }
}

function sendURI(uri) {
    $("#introduce-area").addClass("mdui-hidden");
    log("开始发送请求");
    let okflag = 0;
    for (let i = 0; i < serverURL.length; i++) {
        $.ajax({
            type: "POST",
            dataType: "json",
            url: serverURL[i] + "api/add/uri",
            contentType: "application/json",
            data: JSON.stringify({
                Auth: {
                    Secret: "canoziia",
                },
                URI: uri,
            }),
            success: function (result) {
                if (result.response == 200) {
                    if (okflag == 0) {
                        okflag = 1;
                        log("请求发送成功");
                        log("已添加种子，磁力链接为: " + '<a href="' + result.magnet + '" target="_blank">[磁力链接]</a> ');
                        startMagnet(result.magnet);
                    }
                } else {
                    log("请求发送失败: " + serverURL[i]);
                }
            },
        });
    }
}

function startMagnet(magnet) {
    client.add(
        magnet,
        {
            announce: trackerURL,
        },
        onTorrent
    );
}

function onTorrent(torrent) {
    log("已获取种子信息");
    log("种子名: " + torrent.name);
    log(
        "哈希值: " +
            torrent.infoHash +
            " " +
            //  '<a href="' + torrent.magnetURI + '" target="_blank">[磁力链接]</a> ' +
            '<a href="' +
            torrent.torrentFileBlobURL +
            '" target="_blank" download="' +
            torrent.name +
            '.torrent">[下载 .torrent]</a>'
    );
    // Print out progress every 5 seconds
    const interval = setInterval(function () {
        log("Progress: " + (torrent.progress * 100).toFixed(1) + "%");
    }, 5000);

    torrent.on("done", function () {
        log("Progress: 100%");
        clearInterval(interval);
    });

    // Render all files into to the page
    torrent.files.forEach(function (file) {
        if (isExt(file.name, videoExt)) {
            addList(file.name, function () {
                $("#video-area").removeClass("mdui-hidden");
                file.renderTo("#video");
                $("#video-title").text(file.name);
                $(document).attr("title", file.name);
            });
            file.getBlobURL(function (err, url) {
                if (err) return log(err.message);
                log("文件下载完成。");
                log(`<a target="_blank" download="` + file.name + `" href="` + url + `">保存文件: ` + file.name + `</a>`);
            });
        } else if (isExt(file.name, imageExt)) {
            addList(file.name, function () {
                $("#other-file").empty();
                $("#other-area").removeClass("mdui-hidden");
                file.appendTo("#other-file");
            });
        } else if (isExt(file.name, audioExt)) {
            addList(file.name, function () {
                $("#other-file").empty();
                $("#other-area").removeClass("mdui-hidden");
                file.appendTo("#other-file");
            });
        } else {
            log("不支持播放" + file.name);
        }
    });
    // torrent.files.forEach(function (file) {
    //     file.renderTo("#video");
    //     // log('(Blob URLs only work if the file is loaded from a server. "http//localhost" works. "file://" does not.)');
    //     file.getBlobURL(function (err, url) {
    //         if (err) return log(err.message);
    //         log("文件下载完成。");
    //         log('<a href="' + url + '">保存文件: ' + file.name + "</a>");
    //     });
    // });
}

function addList(filename, clickfunc) {
    $("#file-list-area").removeClass("mdui-hidden");
    let li = $(`<li class="mdui-list-item mdui-ripple">` + filename + `</li>`);
    li.click(clickfunc);
    $("#file-list").append(li);
}

function log(str) {
    let logEle = $("#log");
    let p = $("<p>" + str + "</p>");
    logEle.append(p);
    // logEle.val(logEle.val() + str + "\n");
}
