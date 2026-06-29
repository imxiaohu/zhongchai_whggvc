import { getNewsDetail, downloadAttachment } from '../../api/news.js';
import { showToast, navigateBack } from '../../api/page.js';
import ImagePreview from '../../../components/ImagePreview.vue';
import { BASE_URL } from '../../../utils/request.config.js';

export default {
  components: {
    ImagePreview
  },
  data() {
    return {
      newsId: '',
      newsDetail: null,
      loading: false,
      error: '',
      contentImages: [],
      navPaddingTop: '0px',
      showImagePreview: false,
      previewImageUrl: '',
      statusBarHeight: 20,
    };
  },
  
  onLoad(options) {
    try {
      const systemInfo = uni.getSystemInfoSync();
      this.statusBarHeight = systemInfo.statusBarHeight || 20;
    } catch (e) {
      this.statusBarHeight = 20;
    }

    if (options.id) {
      this.newsId = options.id;
      this.loadNewsDetail();
    } else {
    this.error = '通知ID为空';
    }
  },

  mounted() {
  },

  updated() {
  },

  beforeDestroy() {
    if (typeof document !== 'undefined') {
      document.body.style.overflow = 'auto';
    }
  },
  
  methods: {
    handleNavHeightReady(navInfo) {
      console.log('导航栏高度信息:', navInfo);
      this.navPaddingTop = navInfo.heightPx;
    },
    
    async loadNewsDetail() {
      if (!this.newsId) return;

      console.log('loadNewsDetail: 开始加载新闻详情, ID:', this.newsId);
      this.loading = true;
      this.error = '';

      try {
        const detail = await getNewsDetail(this.newsId);
        console.log('loadNewsDetail: 获取到新闻详情:', detail);

        if (!detail || typeof detail !== 'object') {
          throw new Error('获取详情失败，返回数据异常');
        }

        if (detail.attachments && detail.attachments.length > 0) {
          console.log('loadNewsDetail: 附件数据:', detail.attachments);
          detail.attachments.forEach((attachment, index) => {
            console.log(`附件 ${index}:`, {
              name: attachment.name,
              size: attachment.size,
              sizeType: typeof attachment.size,
              url: attachment.attachmenturl
            });
          });
        }

        this.newsDetail = detail;
        this.loading = false;

        uni.setNavigationBarTitle({
          title: detail.title || '通知详情'
        });

        console.log('loadNewsDetail: 新闻详情加载完成');
      } catch (error) {
        console.error('loadNewsDetail: 获取新闻详情失败:', error);
        this.error = '获取新闻详情失败，请稍后重试';
        this.loading = false;

        showToast({
          title: '获取新闻详情失败',
          icon: 'none'
        });
      }
    },
    
    formatDate(dateStr) {
      if (!dateStr) return '';
      
      try {
        const date = new Date(dateStr);
        return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
      } catch (error) {
        return dateStr;
      }
    },
    
    formatContent(content) {
      if (!content) return '';

      try {
        const tempImages = [];

        let formattedContent = content
          .replace(/\n/g, '<br/>')
          .replace(/\s{2,}/g, '&nbsp;&nbsp;');

        formattedContent = formattedContent.replace(
          /<img([^>]*?)>/gi,
          (match, attributes) => {
            try {
              const srcMatch = /src\s*=\s*["']([^"']*)["']/i.exec(attributes);
              if (!srcMatch) {
                return match;
              }

              const imageUrl = srcMatch[1];
              tempImages.push(imageUrl);
              const imageIndex = tempImages.length - 1;

              const attributeMap = {};

              const attrRegex = /(\w+)\s*=\s*["']([^"']*)["']/g;
              let attrMatch;
              while ((attrMatch = attrRegex.exec(attributes)) !== null) {
                attributeMap[attrMatch[1].toLowerCase()] = attrMatch[2];
              }

              let styles = [];
              if (attributeMap.style) {
                const existingStyles = attributeMap.style
                  .split(';')
                  .map(s => s.trim())
                  .filter(s => s && !s.match(/^(max-width|width|height|display|margin|cursor|border-radius|transition)\s*:/i));

                styles = styles.concat(existingStyles);
              }

              styles.push('max-width: 100%');
              styles.push('height: auto');
              styles.push('display: block');
              styles.push('margin: 10px auto');
              styles.push('cursor: pointer');
              styles.push('border-radius: 8px');
              styles.push('transition: transform 0.2s ease');

              const newAttributes = [];

              newAttributes.push(`src="${imageUrl}"`);

              Object.keys(attributeMap).forEach(key => {
                if (key !== 'style' && key !== 'src') {
                  newAttributes.push(`${key}="${attributeMap[key]}"`);
                }
              });

              newAttributes.push(`style="${styles.join('; ')}"`);

              newAttributes.push(`data-image-index="${imageIndex}"`);
              newAttributes.push(`data-clickable="true"`);

              return `<img ${newAttributes.join(' ')}>`;
            } catch (imgError) {
              console.warn('formatContent: 单个图片处理失败，保持原样:', imgError, match);
              return match;
            }
          }
        );

        if (formattedContent.includes('< ') || formattedContent.includes(' >')) {
          console.warn('formatContent: 检测到可能的HTML语法错误，使用原始内容');
          return content.replace(/\n/g, '<br/>').replace(/\s{2,}/g, '&nbsp;&nbsp;');
        }

        if (JSON.stringify(tempImages) !== JSON.stringify(this.contentImages)) {
          this.contentImages = tempImages;
        }

        return formattedContent;
      } catch (error) {
        console.error('formatContent: 内容格式化失败，使用原始内容:', error);
        return content.replace(/\n/g, '<br/>').replace(/\s{2,}/g, '&nbsp;&nbsp;');
      }
    },
    
    formatFileSize(size) {
      if (!size) return '';

      if (typeof size === 'string' && /^\d+(\.\d+)?(B|KB|MB|GB)$/i.test(size)) {
        return size;
      }

      if (typeof size === 'string') {
        const match = size.match(/^(\d+(?:\.\d+)?)/);
        if (match) {
          size = parseFloat(match[1]);
        } else {
          return size;
        }
      }

      const numSize = Number(size);
      if (isNaN(numSize)) {
        return size.toString();
      }

      const units = ['B', 'KB', 'MB', 'GB'];
      let unitIndex = 0;
      let fileSize = numSize;

      while (fileSize >= 1024 && unitIndex < units.length - 1) {
        fileSize /= 1024;
        unitIndex++;
      }

      return `${fileSize.toFixed(1)}${units[unitIndex]}`;
    },
    
    downloadAttachment(attachment) {
      console.log('downloadAttachment: 准备下载附件:', attachment);

      if (!attachment.attachmenturl) {
        showToast({
          title: '附件链接无效',
          icon: 'none'
        });
        return;
      }

      this.downloadViaProxy(attachment);
    },

    async downloadViaProxy(attachment) {
      console.log('downloadViaProxy: 通过代理下载附件:', attachment);

      const proxyUrl = `${BASE_URL}/api/proxy/download-attachment`;
      
      const attachmentUrl = encodeURIComponent(attachment.attachmenturl);
      const fileName = encodeURIComponent(attachment.name || 'attachment');
      const queryParams = `attachmentUrl=${attachmentUrl}&fileName=${fileName}`;
      
      const downloadUrl = `${proxyUrl}?${queryParams}`;

      console.log('downloadViaProxy: 代理下载链接:', downloadUrl);

      showToast({
        title: '开始下载...',
        icon: 'loading',
        duration: 2000
      });

      try {
        const token = uni.getStorageSync('token');
        if (!token) {
          showToast({
            title: '请先登录',
            icon: 'none'
          });
          return;
        }

        const downloadTask = uni.downloadFile({
          url: downloadUrl,
          header: {
            'Authorization': `Bearer ${token}`,
            'X-Access-Token': token,
            'Accept': 'application/octet-stream, */*'
          },
          success: async (res) => {
            console.log('downloadViaProxy: 下载响应:', res);

            if (res.statusCode === 200) {
              // #region agent log
              fetch('http://127.0.0.1:7819/ingest/3e77e2d8-b579-4c0c-b6a8-82f5adf59d3a', {method:'POST', headers:{'Content-Type':'application/json','X-Debug-Session-Id':'28a407'}, body:JSON.stringify({sessionId:'28a407',location:'detail.js:290',message:'downloadViaProxy: statusCode=200, checking tempFile for JSON error',data:{tempFilePath:res.tempFilePath},timestamp:Date.now()})}).catch(()=>{});
              // #endregion

              // H5: 检查 tempFilePath 是否为 JSON 错误响应
              // #ifdef H5
              if (res.tempFilePath && res.tempFilePath.startsWith('blob:')) {
                try {
                  const textResponse = await fetch(res.tempFilePath);
                  const text = await textResponse.text();
                  // #region agent log
                  fetch('http://127.0.0.1:7819/ingest/3e77e2d8-b579-4c0c-b6a8-82f5adf59d3a', {method:'POST', headers:{'Content-Type':'application/json','X-Debug-Session-Id':'28a407'}, body:JSON.stringify({sessionId:'28a407',location:'detail.js:310',message:'downloadViaProxy: H5 blob content check',data:{textLength:text.length,firstBytes:text.substring(0,50)},timestamp:Date.now()})}).catch(()=>{});
                  // #endregion
                  if (text.startsWith('{') || text.startsWith('[')) {
                    try {
                      const jsonResp = JSON.parse(text);
                      // #region agent log
                      fetch('http://127.0.0.1:7819/ingest/3e77e2d8-b579-4c0c-b6a8-82f5adf59d3a', {method:'POST', headers:{'Content-Type':'application/json','X-Debug-Session-Id':'28a407'}, body:JSON.stringify({sessionId:'28a407',location:'detail.js:323',message:'downloadViaProxy: H5 detected JSON error response from backend',data:{jsonResp},timestamp:Date.now()})}).catch(()=>{});
                      // #endregion
                      console.error('downloadViaProxy: 后端返回JSON错误响应:', jsonResp);
                      if (jsonResp.result && jsonResp.result.needManualDownload) {
                        uni.showModal({
                          title: '需要手动下载',
                          content: `${jsonResp.message || '文件服务器需要特殊认证'}\n\n文件名: ${jsonResp.result.fileName || attachment.name}\n\n是否复制下载链接到剪贴板？`,
                          confirmText: '复制链接',
                          success: (modalRes) => {
                            if (modalRes.confirm) {
                              uni.setClipboardData({
                                data: jsonResp.result.downloadUrl,
                                success: () => uni.showToast({ title: '链接已复制，请在浏览器打开', icon: 'none' })
                              });
                            }
                          }
                        });
                        return;
                      }
                      uni.showToast({
                        title: jsonResp.message || '下载失败',
                        icon: 'none',
                        duration: 3000
                      });
                      return;
                    } catch (parseErr) {
                      // 不是有效的 JSON，继续正常流程
                    }
                  }
                } catch (fetchErr) {
                  console.warn('downloadViaProxy: 无法读取 blob 内容:', fetchErr);
                }
              }
              // #endif

              // #ifndef H5
              // App/小程序：检查 Content-Length 是否过小（JSON 错误体通常 < 1KB）
              const likelyError = res.totalBytesExpectedToWrite > 0 && res.totalBytesExpectedToWrite < 1024;
              // #endif

              console.log('downloadViaProxy: 下载成功，临时文件路径:', res.tempFilePath);

              // #ifdef H5
              this.triggerBrowserDownload(res.tempFilePath, attachment.name);
              showToast({
                title: '下载完成',
                icon: 'success'
              });
              // #endif

              // #ifndef H5
              if (likelyError) {
                uni.getFileInfo({
                  filePath: res.tempFilePath,
                  success: (info) => {
                    if (info.size < 1024) {
                      uni.readFile({
                        filePath: res.tempFilePath,
                        encoding: 'utf8',
                        success: (readRes) => {
                          try {
                            const jsonData = JSON.parse(readRes.data);
                            console.error('downloadViaProxy: 检测到 JSON 错误响应:', jsonData);
                            uni.showToast({
                              title: jsonData.message || '下载失败',
                              icon: 'none'
                            });
                            return;
                          } catch (e) {
                            // 不是 JSON，正常处理
                          }
                          this.saveDownloadedFile(res.tempFilePath, attachment.name);
                        },
                        fail: () => this.saveDownloadedFile(res.tempFilePath, attachment.name)
                      });
                      return;
                    }
                    this.saveDownloadedFile(res.tempFilePath, attachment.name);
                  },
                  fail: () => this.saveDownloadedFile(res.tempFilePath, attachment.name)
                });
                return;
              }
              this.saveDownloadedFile(res.tempFilePath, attachment.name);
              // #endif
            } else {
              console.error('downloadViaProxy: 下载失败，状态码:', res.statusCode);
              this.handleDownloadError(res.statusCode, attachment);
            }
          },
          fail: (error) => {
            console.error('downloadViaProxy: 下载请求失败:', error);
            this.handleDownloadError(null, attachment, error);
          }
        });

        downloadTask.onProgressUpdate((res) => {
          console.log('下载进度:', res.progress + '%');
          console.log('已下载:', res.totalBytesWritten);
          console.log('总大小:', res.totalBytesExpectedToWrite);
        });

      } catch (error) {
        console.error('downloadViaProxy: 下载异常:', error);
        this.handleDownloadError(null, attachment, error);
      }
    },

    handleDownloadError(statusCode, attachment, error) {
      let errorMessage = '下载失败';

      if (statusCode === 401) {
        errorMessage = '认证失败，请重新登录';
      } else if (statusCode === 404) {
        errorMessage = '文件不存在';
      } else if (statusCode === 403) {
        errorMessage = '没有下载权限';
      } else if (error && error.errMsg) {
        errorMessage = error.errMsg;
      }

      showToast({
        title: errorMessage,
        icon: 'none'
      });

      setTimeout(() => {
        this.showDownloadOptions(attachment);
      }, 1500);
    },

    triggerBrowserDownload(tempFilePath, fileName) {
      try {
        const link = document.createElement('a');
        link.href = tempFilePath;
        link.download = fileName || 'attachment';
        link.style.display = 'none';

        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);

        console.log('triggerBrowserDownload: 已触发浏览器下载');
      } catch (error) {
        console.error('triggerBrowserDownload: 触发浏览器下载失败:', error);
      }
    },

    // #ifndef H5
    saveDownloadedFile(tempFilePath, fileName) {
      uni.saveFile({
        tempFilePath: tempFilePath,
        success: (saveRes) => {
          console.log('saveDownloadedFile: 文件保存成功:', saveRes);
          showToast({
            title: '下载完成',
            icon: 'success'
          });
        },
        fail: (saveError) => {
          console.error('saveDownloadedFile: 文件保存失败:', saveError);
          showToast({
            title: '文件保存失败',
            icon: 'none'
          });
        }
      });
    },
    // #endif

    showDownloadOptions(attachment) {
      console.log('showDownloadOptions: 显示下载选项');

      const possibleUrls = [
        `https://scs.whggvc.net/scsoa/sys/common/static/${attachment.attachmenturl}`,
        `https://scs.whggvc.net/scloudoa/sys/common/static/${attachment.attachmenturl}`,
        `https://scs.whggvc.net/static/${attachment.attachmenturl}`
      ];

      uni.showActionSheet({
        itemList: [
          '在新窗口打开',
          '复制下载链接',
          '取消'
        ],
        success: (res) => {
          if (res.tapIndex === 0) {
            this.openInNewWindow(possibleUrls, attachment);
          } else if (res.tapIndex === 1) {
            this.copyToClipboard(possibleUrls[0], attachment.name);
          }
        }
      });
    },

    openInNewWindow(urls, attachment) {
      console.log('openInNewWindow: 尝试在新窗口打开文件');

      try {
        window.open(urls[0], '_blank');
        showToast({
          title: '已在新窗口打开',
          icon: 'success'
        });
      } catch (error) {
        console.error('openInNewWindow: 打开新窗口失败:', error);
        this.showAllDownloadLinks(urls, attachment);
      }
    },

    showAllDownloadLinks(urls, attachment) {
      const linkList = urls.map((url, index) => `链接${index + 1}`);
      linkList.push('取消');

      uni.showActionSheet({
        itemList: linkList,
        success: (res) => {
          if (res.tapIndex < urls.length) {
            this.copyToClipboard(urls[res.tapIndex], attachment.name);
          }
        }
      });
    },

    downloadInBrowser(attachment) {
      const possibleBaseUrls = [
        'https://scs.whggvc.net/scsoa/sys/common/static/',
        'https://scs.whggvc.net/scloudoa/sys/common/static/',
        'https://scs.whggvc.net/static/',
        ''
      ];

      this.tryBrowserDownload(attachment, possibleBaseUrls, 0);
    },

    async tryBrowserDownload(attachment, baseUrls, urlIndex) {
      if (urlIndex >= baseUrls.length) {
        this.showDownloadOptions(attachment);
        return;
      }

      const baseUrl = baseUrls[urlIndex];
      let downloadUrl = attachment.attachmenturl;

      if (!attachment.attachmenturl.startsWith('http')) {
        downloadUrl = baseUrl + attachment.attachmenturl;
      }

      console.log(`tryBrowserDownload: 尝试下载链接 ${urlIndex}:`, downloadUrl);

      try {
        const response = await fetch(downloadUrl, { method: 'HEAD' });

        if (response.ok) {
          const link = document.createElement('a');
          link.href = downloadUrl;
          link.download = attachment.name || 'attachment';
          link.target = '_blank';
          link.style.display = 'none';

          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);

          showToast({
            title: '开始下载...',
            icon: 'success'
          });
          return;
        } else {
          throw new Error(`HTTP ${response.status}`);
        }
      } catch (error) {
        console.warn(`tryBrowserDownload: 链接 ${urlIndex} 失败:`, error.message);
        this.tryBrowserDownload(attachment, baseUrls, urlIndex + 1);
      }
    },

    copyToClipboard(url, fileName) {
      try {
        if (navigator.clipboard && window.isSecureContext) {
          navigator.clipboard.writeText(url).then(() => {
            showToast({
              title: '链接已复制到剪贴板',
              icon: 'success'
            });
          }).catch(() => {
            this.fallbackCopyToClipboard(url, fileName);
          });
        } else {
          this.fallbackCopyToClipboard(url, fileName);
        }
      } catch (error) {
        console.error('copyToClipboard: 复制失败:', error);
        this.fallbackCopyToClipboard(url, fileName);
      }
    },

    fallbackCopyToClipboard(url, fileName) {
      try {
        const textArea = document.createElement('textarea');
        textArea.value = url;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        const successful = document.execCommand('copy');
        document.body.removeChild(textArea);

        if (successful) {
          showToast({
            title: '链接已复制到剪贴板',
            icon: 'success'
          });
        } else {
          throw new Error('execCommand failed');
        }
      } catch (error) {
        console.error('fallbackCopyToClipboard: 备用复制失败:', error);
        uni.showModal({
          title: '下载链接',
          content: `请手动复制以下链接下载文件：\n${url}`,
          showCancel: false,
          confirmText: '知道了'
        });
      }
    },

    downloadWithUniAPI(attachment) {
      const possibleBaseUrls = [
        'https://scs.whggvc.net/scsoa/sys/common/static/',
        'https://scs.whggvc.net/scloudoa/sys/common/static/',
        'https://scs.whggvc.net/static/',
        ''
      ];

      let downloadUrl = attachment.attachmenturl;

      if (!attachment.attachmenturl.startsWith('http')) {
        downloadUrl = possibleBaseUrls[0] + attachment.attachmenturl;
      }

      console.log('downloadWithUniAPI: 下载链接:', downloadUrl);

      showToast({
        title: '开始下载...',
        icon: 'loading',
        duration: 1000
      });

      uni.downloadFile({
        url: downloadUrl,
        success: (res) => {
          console.log('downloadWithUniAPI: 下载响应:', res);

          if (res.statusCode === 200) {
            showToast({
              title: '下载完成',
              icon: 'success'
            });

            uni.saveFile({
              tempFilePath: res.tempFilePath,
              success: (saveRes) => {
                console.log('downloadWithUniAPI: 文件保存成功:', saveRes);
                showToast({
                  title: '文件已保存',
                  icon: 'success'
                });
              },
              fail: (saveError) => {
                console.error('downloadWithUniAPI: 文件保存失败:', saveError);
                showToast({
                  title: '文件保存失败',
                  icon: 'none'
                });
              }
            });
          } else if (res.statusCode === 404) {
            console.warn('downloadWithUniAPI: 文件不存在，尝试其他URL');
            this.tryAlternativeDownload(attachment, possibleBaseUrls, 1);
          } else {
            console.error('downloadWithUniAPI: 下载失败，状态码:', res.statusCode);
            showToast({
              title: `下载失败 (${res.statusCode})`,
              icon: 'none'
            });
          }
        },
        fail: (error) => {
          console.error('downloadWithUniAPI: 下载失败:', error);
          showToast({
            title: '网络错误，下载失败',
            icon: 'none'
          });
        }
      });
    },

    tryAlternativeDownload(attachment, baseUrls, urlIndex) {
      if (urlIndex >= baseUrls.length) {
        showToast({
          title: '所有下载链接都无效',
          icon: 'none'
        });
        return;
      }

      const baseUrl = baseUrls[urlIndex];
      const downloadUrl = baseUrl + attachment.attachmenturl;

      console.log(`tryAlternativeDownload: 尝试备用链接 ${urlIndex}:`, downloadUrl);

      uni.downloadFile({
        url: downloadUrl,
        success: (res) => {
          if (res.statusCode === 200) {
            showToast({
              title: '下载完成',
              icon: 'success'
            });

            uni.saveFile({
              tempFilePath: res.tempFilePath,
              success: (saveRes) => {
                console.log('tryAlternativeDownload: 文件保存成功:', saveRes);
                showToast({
                  title: '文件已保存',
                  icon: 'success'
                });
              },
              fail: (saveError) => {
                console.error('tryAlternativeDownload: 文件保存失败:', saveError);
                showToast({
                  title: '文件保存失败',
                  icon: 'none'
                });
              }
            });
          } else {
            this.tryAlternativeDownload(attachment, baseUrls, urlIndex + 1);
          }
        },
        fail: (error) => {
          console.error(`tryAlternativeDownload: 备用链接 ${urlIndex} 失败:`, error);
          this.tryAlternativeDownload(attachment, baseUrls, urlIndex + 1);
        }
      });
    },
    
    shareNews() {
      if (!this.newsDetail) return;
      
      uni.share({
        provider: 'weixin',
        scene: 'WXSceneSession',
        type: 0,
        href: '',
        title: this.newsDetail.title,
        summary: this.newsDetail.summary || this.newsDetail.title,
        imageUrl: this.newsDetail.cover || '',
        success: () => {
          showToast({
            title: '分享成功',
            icon: 'success'
          });
        },
        fail: (error) => {
          console.error('分享失败:', error);
          showToast({
            title: '分享失败',
            icon: 'none'
          });
        }
      });
    },
    
    goBack() {
      navigateBack();
    },

    // #ifdef H5
    bindImageClickEvents() {
      this.$nextTick(() => {
        const richTextContainer = document.querySelector('.detail-body rich-text, .detail-body');
        if (!richTextContainer) return;
        const images = richTextContainer.querySelectorAll('img');
        images.forEach(img => {
          img.style.cursor = 'pointer';
          img.onclick = (e) => {
            e.stopPropagation();
            const src = img.src || img.getAttribute('src');
            if (src) {
              uni.previewImage({
                current: src,
                urls: this.contentImages.length > 0 ? this.contentImages : [src]
              });
            }
          };
        });
      });
    },
    // #endif

    handleRichTextTap(event) {
      console.log('handleRichTextTap: 富文本点击事件:', event);
      console.log('handleRichTextTap: 事件详情:', JSON.stringify(event, null, 2));

      // #ifdef H5
      this.bindImageClickEvents();
      // #endif

      // #ifdef MP-WEIXIN
      console.log('handleRichTextTap: 微信小程序环境处理');
      
      if (event && event.detail && event.detail.node) {
        const node = event.detail.node;
        console.log('handleRichTextTap: 微信小程序节点信息 (方法1):', node);
        
        if (node.name === 'img' && node.attrs && node.attrs.src) {
          const imageUrl = node.attrs.src;
          const imageIndex = node.attrs['data-image-index'];
          
          console.log('handleRichTextTap: 微信小程序图片点击 (方法1):', {
            imageUrl,
            imageIndex,
            attrs: node.attrs
          });
          
          this.previewImage(imageUrl);
          return;
        }
      }
      
      if (event && (event.target || event.currentTarget)) {
        const target = event.target || event.currentTarget;
        console.log('handleRichTextTap: 微信小程序目标信息 (方法2):', target);
        
        if (target.dataset) {
          console.log('handleRichTextTap: 目标数据集:', target.dataset);
          
          if (target.dataset.imageIndex !== undefined && this.contentImages[target.dataset.imageIndex]) {
            const imageUrl = this.contentImages[target.dataset.imageIndex];
            console.log('handleRichTextTap: 微信小程序图片点击 (方法2):', {
              imageUrl,
              imageIndex: target.dataset.imageIndex
            });
            
            this.previewImage(imageUrl);
            return;
          }
        }
      }
      
      if (event && event.detail) {
        console.log('handleRichTextTap: 微信小程序事件详情 (方法3):', event.detail);
        
        if (event.detail.target && event.detail.target.dataset) {
          const dataset = event.detail.target.dataset;
          console.log('handleRichTextTap: 详情目标数据集:', dataset);
          
          if (dataset.imageIndex !== undefined && this.contentImages[dataset.imageIndex]) {
            const imageUrl = this.contentImages[dataset.imageIndex];
            console.log('handleRichTextTap: 微信小程序图片点击 (方法3):', {
              imageUrl,
              imageIndex: dataset.imageIndex
            });
            
            this.previewImage(imageUrl);
            return;
          }
        }
      }
      
      if (event && event.detail && event.detail.x !== undefined && event.detail.y !== undefined) {
        console.log('handleRichTextTap: 尝试通过坐标识别图片点击:', {
          x: event.detail.x,
          y: event.detail.y,
          contentImages: this.contentImages
        });
        
        if (this.contentImages && this.contentImages.length === 1) {
          console.log('handleRichTextTap: 检测到单张图片，直接预览');
          this.previewImage(this.contentImages[0]);
          return;
        }
        
        if (this.contentImages && this.contentImages.length > 0) {
          console.log('handleRichTextTap: 检测到多张图片，预览第一张');
          this.previewImage(this.contentImages[0]);
          return;
        }
      }
      
      console.log('handleRichTextTap: 微信小程序未识别到图片点击');
      // #endif

      // #ifdef MP-ALIPAY || MP-BAIDU || MP-TOUTIAO || MP-QQ
      if (event && event.detail) {
        console.log('handleRichTextTap: 其他小程序平台事件:', event.detail);
      }
      // #endif
    },

    closeImagePreview() {
      console.log('closeImagePreview: 关闭图片预览');
      this.showImagePreview = false;
      this.previewImageUrl = '';
    },

    previewImage(imageUrl) {
      console.log('previewImage: 预览图片:', imageUrl);
      this.previewImageUrl = imageUrl;
      this.showImagePreview = true;
    },
  }
};
