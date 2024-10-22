# rod-crawl-learn
基于go-rod离线 `learn.lianglianglee.com` 技术摘抄

功能：
1. 爬取`learn.lianglianglee.com`[专栏](https://learn.lianglianglee.com/%e4%b8%93%e6%a0%8f)，以获取到各个专栏目录页面的`url`
2. 逐个爬取专栏目录页面，获取到每个目录下的所有文章的`url`
3. 爬取文章页面，获取到文章内容，并保存到本地`mhtml`文件，文件目录按专栏目录，文件名按文章标题命名
4. 爬取的各个专栏有状态: `0:未爬取, 100:正在爬取, 200:已爬取完, 400:爬取出错`
5. 爬取的每个文章有状态: `0:未爬取, 100:正在爬取, 200:已爬取完, 400:爬取出错`

因为`https://learn.lianglianglee.com/`有反爬机制，所以目前简单实现单线程有节制爬取