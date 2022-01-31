const videoExt = ["mp4", "ogg", "webm"];

const audioExt = ["mp3", "wav", "flac", "m3u"];

const imageExt = ["png", "jpg", "jpeg", "webp"];

const serverURL = ["https://wt-server.dnlab.net/1/", "https://wt-server.dnlab.net/2/", "https://wt-server.dnlab.net/3/", "/"];

const trackerURL = ["wss://tracker.lili.ac", "wss://tracker.btorrent.xyz", "wss://tracker.openwebtorrent.com"];

const loopTime = 2000;

const webseedPrefix = "https://webseed.lili.ac/";

const wenseedSuffix = "?hash=k5Znwdx3&download=1";

const client = new WebTorrent();

const md = markdownit();

var printInfo;

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
    $("#upload-filename").text($("#upload-torrent")[0].files[0].name);
});

$("#upload-submit").click(function (e) {
    if ($("#upload-torrent")[0].files[0] == undefined) {
        log("还没选择要上传的文件!");
        return;
    }
    uploadTorrent();
});

$("#share-button").click(function (e) {
    let urlsuffix = "/?";
    client.torrents.forEach(function (t) {
        urlsuffix += `uri=${uriEncode(t.magnetURI)}&`;
    });
    urlsuffix = urlsuffix.substr(0, urlsuffix.length - 1);
    let res = window.location.origin + urlsuffix;
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

function initPage() {
    loadMarkdown();
    loadShareURL();
}

// 检测url并加载链接
function loadShareURL() {
    let mList = getQueryVariable("uri");
    if (mList.length == 0) {
        return;
    }
    mList.forEach(function (strin) {
        sendURI(uriDecode(strin));
    });
}

function loadMarkdown() {
    $("[mdtarget]").each(function () {
        let e = $(this);
        if (e.attr("mdsource") != typeof undefined) {
            $.get(e.attr("mdsource"), function (result) {
                $(e.attr("mdtarget"))
                    .empty()
                    .append($(md.render(result)));
            });
        } else {
            $(e.attr("mdtarget"))
                .empty()
                .append($(md.render(e.text())));
        }
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
    log("开始上传文件");
    let okflag = 0;
    let formData = new FormData();
    formData.append("torrent", $("#upload-torrent")[0].files[0]);
    for (let i = 0; i < serverURL.length; i++) {
        $.ajax({
            url: serverURL[i] + "api/add/torrent",
            type: "POST",
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
                    log("torrent 上传失败[200]: " + serverURL[i]);
                }
            },
            error: function (e) {
                if (okflag == 0) {
                    log("torrent 上传失败: " + serverURL[i]);
                }
            },
        });
    }
}

function sendURI(uri) {
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
                    log("请求发送失败[200]: " + serverURL[i]);
                }
            },
            error: function (e) {
                if (okflag == 0) {
                    log("请求发送失败: " + serverURL[i]);
                }
            },
        });
    }
}

function startMagnet(magnet) {
    // client.add(
    //     magnet,
    //     {
    //         announce: trackerURL,
    //     },
    //     onTorrent
    // );
    client.add(magnet, onTorrent);
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
    // const interval = setInterval(function () {
    //     log("Progress: " + (torrent.progress * 100).toFixed(1) + "%");
    // }, 5000);

    // torrent.on("done", function () {
    //     log("Progress: 100%");
    //     clearInterval(interval);
    // });

    // Render all files into to the page
    torrent.files.forEach(function (file) {
        if (isExt(file.name, videoExt)) {
            addList(file, function () {
                $("#video-area").removeClass("mdui-hidden");
                file.renderTo("#video");
                $("#video").ready(function (e) {
                    $("#video").css("aspect-ratio", "");
                });
                $("#video-title").text(file.name);
                $(document).attr("title", file.name);
            });
            file.getBlobURL(function (err, url) {
                if (err) return log(err.message);
                log("文件下载完成。");
                log(`<a target="_blank" download="` + file.name + `" href="` + url + `">保存文件: ` + file.name + `</a>`);
            });
        } else if (isExt(file.name, imageExt)) {
            addList(file, function () {
                $("#other-file").empty();
                $("#other-area").removeClass("mdui-hidden");
                file.appendTo("#other-file");
            });
        } else if (isExt(file.name, audioExt)) {
            addList(file, function () {
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

function filesize(s) {
    var res = "";
    if (s < 1000) {
        //如果小于1KB转化成B
        res = s.toPrecision(3) + "B";
    } else if (s < 1000 * 1000) {
        //如果小于1MB转化成KB
        res = (s / 1000).toPrecision(3) + "K";
    } else if (s < 1000 * 1000 * 1000) {
        //如果小于1GB转化成MB
        res = (s / (1000 * 1000)).toPrecision(3) + "M";
    } else {
        //其他转化成GB
        res = (s / (1000 * 1000 * 1000)).toPrecision(3) + "G";
    }
    return res;
}

function getDownloadInfo(file) {
    return `${filesize(file._torrent.downloadSpeed)}/S (${filesize(file._torrent.downloaded)})`;
}

function getUploadInfo(file) {
    return `${filesize(file._torrent.uploadSpeed)}/S (${filesize(file._torrent.uploaded)})`;
}

function getPeerInfo(file) {
    return file._torrent.wires.length;
}

function getProgressInfo(file) {
    return file.progress.toPrecision(2);
}

function getFileTorrentHash(file) {
    return file._torrent.infoHash;
}

function addList(file, clickfunc) {
    $("#file-list-area").removeClass("mdui-hidden");
    let li = $(`<li class="mdui-list-item mdui-ripple">` + file.name + `</li>`);
    li.click(function () {
        clearInterval(printInfo);
        printInfo = setInterval(function () {
            $("#download-info").text(getDownloadInfo(file));
            $("#upload-info").text(getUploadInfo(file));
            $("#peer-info").text(getPeerInfo(file));
            $("#progress-info").text(getProgressInfo(file));
        }, loopTime);
        $("#delete-torrent")
            .unbind("click")
            .click(function () {
                deleteTorrent(getFileTorrentHash(file));
            });
        $("#peer-info")
            .parent()
            .unbind("click")
            .click(function () {
                let peerip = $("#peer-ip");
                peerip.empty();
                file._torrent.wires.forEach(function (wire) {
                    if (wire.type == "webrtc") {
                        let li = $(`<li class="mdui-menu-item">
                            <a class="mdui-ripple" style="overflow-x: auto">${wire.remoteAddress}</a>
                        </li>`);
                        peerip.append(li);
                    }
                });
                file._torrent.urlList.forEach(function (peerurl) {
                    let li = $(`<li class="mdui-menu-item">
                        <a class="mdui-ripple" style="overflow-x: auto">${peerurl}</a>
                    </li>`);
                    peerip.append(li);
                });
            });
        $("#enable-webseed")
            .unbind("click")
            .click(function (e) {
                if (file._torrent.files.length == 1) {
                    let webseedURL = webseedPrefix + encodeURIComponent(file.path) + wenseedSuffix;
                    file._torrent.addWebSeed(webseedURL);
                    log("已启用webseed");
                } else {
                    log("多文件种子暂不支持webseed");
                }
            });
        $("#disable-webseed")
            .unbind("click")
            .click(function (e) {
                file._torrent.wires.forEach(function (wire) {
                    if (wire.type == "webSeed") {
                        wire.destroy();
                    }
                });
                log("已禁用webseed");
            });
        clickfunc();
    });
    $("#file-list").append(li);
}

function deleteTorrent(hash) {
    log("请求删除种子");
    for (let i = 0; i < serverURL.length; i++) {
        $.ajax({
            type: "POST",
            dataType: "json",
            url: `${serverURL[i]}api/torrent/${hash}/delete`,
            contentType: "application/json",
            data: JSON.stringify({
                Auth: {
                    Secret: "canoziia",
                },
                DeleteFile: "yes",
            }),
            success: function (result) {
                if (result.response == 200) {
                    log("请求删除成功: " + serverURL[i]);
                } else {
                    log("请求删除失败[200]: " + serverURL[i]);
                }
            },
            error: function (e) {
                log("请求删除失败: " + serverURL[i]);
            },
        });
    }
}

function log(str) {
    let logEle = $("#log");
    let p = $("<p>" + str + "</p>");
    logEle.append(p);
    let logParent = logEle.parent();
    logParent.scrollTop(logParent[0].scrollHeight);
    // logEle.val(logEle.val() + str + "\n");
}
