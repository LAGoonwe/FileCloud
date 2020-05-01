/**
 * 自定义分页（封装实现）
 * @type {Page}
 * @data：2019-06-02
 * @author：lzw
 */

P = new Page();
//分页对象
function Page(){
    this.config = {elemId:'#page',pageIndex:'1',total:'0',pageNum:'7',pageSize:'10'};//默认参数
    this.version = '1.0';//分页版本
    this.requestFunction = null;//分页版本

    //初始化参数
    this.initMathod = function(obj){
        $.extend(this.config,obj.params);//默认参数 + 用户自定义参数
        this.requestFunction = obj.requestFunction;
        this.renderPage();
    };

    //渲染分页
    this.renderPage = function(){
        this.requestFunction();
        this.pageHtml();

        //分页绑定事件
        $(P.config.elemId).on('click','a',function(){
            var flag = $(this).parent().hasClass('disabled');
            if(flag){
                return false;
            }

            var pageIndex = $(this).data('pageindex');
            P.config.pageIndex = pageIndex;
            P.requestFunction();
            P.pageHtml();
        });
    };

    //分页合成
    this.pageHtml = function(){
        var data = this.config;
        if(parseInt(data.total) <= 0){
            return false;
        }

        var elemId = data.elemId;
        var pageNum = isBlank(data.pageNum) ? 7 : parseInt(data.pageNum);//可显示页码个数
        var pageSize = isBlank(data.pageSize) ? 10 : parseInt(data.pageSize);//可显示页码个数
        var total = parseInt(data.total);//总记录数
        var pageTotal = total%pageSize != 0 ? parseInt(total/pageSize) + 1 : parseInt(total/pageSize);//总页数
        var pageIndex = pageTotal < parseInt(data.pageIndex) ? pageTotal : parseInt(data.pageIndex);//当前页
        var j = pageTotal < pageNum ? pageTotal : pageNum;//如果总页数小于可见页码，则显示页码为总页数
        var k = pageIndex < parseInt((j/2) + 1) ? -1 * (pageIndex - 1) : pageIndex > (pageTotal - parseInt(j/2)) ? -1 * (j - (pageTotal - pageIndex) - 1) : -1 * parseInt((j/2));//遍历初始值
        var pageHtml = '<ul>';

        if(pageIndex <= 0 || pageIndex == 1){
            pageHtml += '<li class="disabled"><a href="javascript:;" data-pageindex="'+ pageIndex +'">首页</a></li>' +
                '<li class="disabled"><a href="javascript:;" data-pageindex="'+ pageIndex +'">上一页</a></li>';
        }else{
            pageHtml += '<li><a href="javascript:;" data-pageindex="1">首页</a></li>' +
                '<li><a href="javascript:;" data-pageindex="'+ (pageIndex - 1) +'">上一页</a></li>';
        }

        for(var i = k;i < (k + j);i++){
            if(pageTotal == (pageIndex + i - 1))break;
            if(i == 0){
                pageHtml += '<li class="active"><a href="javascript:;" data-pageindex="'+ pageIndex +'">'+ pageIndex +'</a></li>';
            }else{
                pageHtml += '<li><a href="javascript:;" data-pageindex="'+ (pageIndex + i) +'">'+ (pageIndex + i) +'</a></li>';
            }
        }

        if(pageTotal == 1 ||  pageTotal <= pageIndex){
            pageHtml += '<li class="disabled"><a href="javascript:;" data-pageindex="'+ pageTotal +'">下一页</a></li>' +
                '<li class="disabled"><a href="javascript:;" data-pageindex="'+ pageTotal +'">末页</a></li>';
        }else{
            pageHtml += '<li><a href="javascript:;" data-pageindex="'+ (pageIndex + 1) +'">下一页</a></li>' +
                '<li><a href="javascript:;" data-pageindex="'+ pageTotal +'">末页</a></li>';
        }
        pageHtml += '</ul>'
        $(elemId).html('');
        $(elemId).html(pageHtml);
    };
}

function isBlank(str){
    if(str == undefined || str == null){
        return true;
    }
    return false;
}
