# 七牛云图片上传配置指南

## 概述
本项目使用七牛云对象存储服务进行图片上传和存储。支持服务器代理上传和前端直传两种方式。

## 🔧 **服务器端配置**

### **1. 环境变量配置**
在服务器的 `.env` 文件中添加七牛云配置：

```bash
# 七牛云配置
QINIU_ACCESS_KEY=your_access_key_here
QINIU_SECRET_KEY=your_secret_key_here
QINIU_BUCKET=your_bucket_name
QINIU_DOMAIN=your_custom_domain.com
QINIU_ZONE=huadong
QINIU_USE_HTTPS=true
QINIU_MAX_SIZE=10485760
```

### **2. 服务器端上传API实现**
```javascript
// /api/upload/image 端点实现示例
const qiniu = require('qiniu');

// 配置七牛云
const accessKey = process.env.QINIU_ACCESS_KEY;
const secretKey = process.env.QINIU_SECRET_KEY;
const bucket = process.env.QINIU_BUCKET;
const domain = process.env.QINIU_DOMAIN;

const mac = new qiniu.auth.digest.Mac(accessKey, secretKey);
const config = new qiniu.conf.Config();
config.zone = qiniu.zone.Zone_z0; // 华东区域

app.post('/api/upload/image', upload.single('file'), async (req, res) => {
  try {
    const file = req.file;
    if (!file) {
      return res.json({ success: false, message: '没有上传文件' });
    }

    // 生成唯一文件名
    const key = `community/images/${Date.now()}_${Math.random().toString(36).substr(2, 9)}.jpg`;
    
    // 上传到七牛云
    const formUploader = new qiniu.form_up.FormUploader(config);
    const putExtra = new qiniu.form_up.PutExtra();
    const uploadToken = new qiniu.rs.PutPolicy({
      scope: bucket,
      expires: 7200 // 2小时过期
    }).uploadToken(mac);

    formUploader.putFile(uploadToken, key, file.path, putExtra, (err, body, info) => {
      if (err) {
        return res.json({ success: false, message: '上传到七牛云失败', error: err });
      }

      if (info.statusCode === 200) {
        const imageUrl = `https://${domain}/${body.key}`;
        res.json({
          success: true,
          result: {
            url: imageUrl,
            filename: body.key,
            size: file.size,
            type: path.extname(file.originalname),
            hash: body.hash,
            storage: 'qiniu'
          }
        });
      } else {
        res.json({ success: false, message: '上传失败', error: body });
      }
    });
  } catch (error) {
    res.json({ success: false, message: '服务器错误', error: error.message });
  }
});
```

### **3. 获取上传凭证API（用于前端直传）**
```javascript
// /api/upload/token 端点实现
app.get('/api/upload/token', authenticateToken, (req, res) => {
  try {
    const keyPrefix = `community/images/${req.user.id}_${Date.now()}_`;
    
    const putPolicy = new qiniu.rs.PutPolicy({
      scope: bucket,
      expires: 3600, // 1小时过期
      insertOnly: 1,
      fsizeLimit: 10485760, // 10MB限制
      mimeLimit: 'image/*'
    });

    const uploadToken = putPolicy.uploadToken(mac);

    res.json({
      success: true,
      result: {
        token: uploadToken,
        keyPrefix: keyPrefix,
        bucket: bucket,
        domain: domain,
        maxSize: '10MB',
        allowedTypes: ['jpg', 'jpeg', 'png', 'gif', 'webp']
      }
    });
  } catch (error) {
    res.json({ success: false, message: '获取上传凭证失败', error: error.message });
  }
});
```

## 📱 **前端实现**

### **1. 服务器代理上传（推荐）**
```javascript
// 通过服务器代理上传到七牛云
async uploadToQiniuViaServer(filePath) {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: API_BASE_URL + '/api/upload/image',
      filePath: filePath,
      name: 'file',
      header: {
        'Authorization': 'Bearer ' + uni.getStorageSync('token')
      },
      success: (res) => {
        const result = JSON.parse(res.data);
        if (result.success) {
          resolve(result.result.url); // 七牛云图片URL
        } else {
          reject(new Error(result.message));
        }
      },
      fail: reject
    });
  });
}
```

### **2. 前端直传到七牛云**
```javascript
// 前端直传到七牛云
async uploadToQiniuDirectly(filePaths) {
  // 1. 获取上传凭证
  const tokenResponse = await uni.request({
    url: API_BASE_URL + '/api/upload/token',
    method: 'GET',
    header: {
      'Authorization': 'Bearer ' + uni.getStorageSync('token')
    }
  });

  const { token, keyPrefix, domain } = tokenResponse.data.result;

  // 2. 直接上传到七牛云
  const uploadPromises = filePaths.map((filePath, index) => {
    return new Promise((resolve, reject) => {
      const key = keyPrefix + index + '.jpg';
      
      uni.uploadFile({
        url: 'https://upload.qiniup.com', // 七牛云上传域名
        filePath: filePath,
        name: 'file',
        formData: {
          'token': token,
          'key': key
        },
        success: (res) => {
          const result = JSON.parse(res.data);
          const imageUrl = `https://${domain}/${result.key}`;
          resolve(imageUrl);
        },
        fail: reject
      });
    });
  });

  return await Promise.all(uploadPromises);
}
```

## 🔍 **七牛云控制台配置**

### **1. 创建存储空间**
1. 登录七牛云控制台
2. 进入对象存储 → 空间管理
3. 创建新的存储空间
4. 选择合适的存储区域（建议华东）
5. 设置访问控制为"公开"

### **2. 配置自定义域名**
1. 在存储空间设置中添加自定义域名
2. 配置CNAME解析
3. 开启HTTPS（推荐）
4. 设置防盗链（可选）

### **3. 获取密钥**
1. 进入个人中心 → 密钥管理
2. 创建或查看AccessKey和SecretKey
3. 将密钥配置到服务器环境变量中

## 🧪 **测试验证**

### **1. 上传功能测试**
```bash
# 测试服务器代理上传
curl -X POST \
  http://localhost:8080/api/upload/image \
  -H 'Authorization: Bearer your_token' \
  -F 'file=@test.jpg'

# 测试获取上传凭证
curl -X GET \
  http://localhost:8080/api/upload/token \
  -H 'Authorization: Bearer your_token'
```

### **2. 图片访问测试**
```bash
# 测试图片是否可以正常访问
curl -I https://your-domain.com/community/images/test.jpg
```

## 🚀 **性能优化建议**

### **1. 图片压缩**
- 前端上传前进行图片压缩
- 服务器端使用七牛云图片处理服务
- 设置合理的图片质量和尺寸限制

### **2. CDN加速**
- 使用七牛云CDN加速图片访问
- 配置合适的缓存策略
- 启用Gzip压缩

### **3. 安全设置**
- 设置上传文件大小限制
- 限制上传文件类型
- 配置防盗链保护
- 定期清理无用图片

## 📋 **常见问题**

### **Q: 上传失败，提示"上传到七牛云失败"**
A: 检查七牛云配置是否正确，确认AccessKey、SecretKey、Bucket名称等配置无误。

### **Q: 图片上传成功但无法访问**
A: 检查自定义域名配置，确认域名解析正确，存储空间访问权限设置为公开。

### **Q: 前端直传失败**
A: 检查上传凭证是否有效，确认七牛云上传域名是否正确，检查跨域设置。

### **Q: 图片加载慢**
A: 启用CDN加速，优化图片格式和大小，使用适当的图片处理参数。

通过以上配置，您的项目就可以正常使用七牛云进行图片上传和存储了。
