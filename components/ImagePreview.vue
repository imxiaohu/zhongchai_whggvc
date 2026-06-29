<template>
  <view v-if="visible" class="image-preview-overlay" @tap="handleClose">
    <view class="image-preview-container" @tap.stop>
      <!-- #ifdef MP-WEIXIN -->
      <view class="mp-image-wrapper">
        <image
          :src="imageUrl"
          class="mp-preview-image"
          mode="aspectFit"
          @tap="handleMpTap"
          @longpress="handleLongPress"
          :style="{
            transform: `scale(${imageScale}) translate(${imageTranslateX}px, ${imageTranslateY}px)`
          }"
          @touchstart="onTouchStart"
          @touchmove="onTouchMove"
          @touchend="onTouchEnd"
        />
      </view>
      <!-- #endif -->

      <!-- #ifndef MP-WEIXIN -->
      <view
        class="image-wrapper"
        @touchstart="onTouchStart"
        @touchmove="onTouchMove"
        @touchend="onTouchEnd"
        @mousedown="onMouseDown"
        @mousemove="onMouseMove"
        @mouseup="onMouseUp"
        @wheel.prevent="onWheel"
      >
        <image
          :src="imageUrl"
          class="preview-image"
          mode="aspectFit"
          @tap="handleTap"
          :style="{
            transform: `scale(${imageScale}) translate(${imageTranslateX}px, ${imageTranslateY}px)`,
            cursor: imageScale > 1 ? (isDragging ? 'grabbing' : 'grab') : 'zoom-in'
          }"
        />
      </view>
      <!-- #endif -->

      <view
        class="preview-close-btn"
        @tap="handleClose"
        @click="handleClose"
        @touchstart="onCloseTouch"
      >
        <text class="close-icon">×</text>
      </view>
      <view class="preview-controls">
        <view class="control-btn" @tap="onZoomOut">
          <text class="control-icon">-</text>
        </view>
        <view class="zoom-info">
          <text>{{Math.round(imageScale * 100)}}%</text>
        </view>
        <view class="control-btn" @tap="onZoomIn">
          <text class="control-icon">+</text>
        </view>
        <view class="control-btn" @tap="onReset">
          <text class="control-text">重置</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
export default {
  name: 'ImagePreview',

  props: {
    visible: {
      type: Boolean,
      default: false
    },
    imageUrl: {
      type: String,
      default: ''
    }
  },

  data() {
    return {
      imageScale: 1,
      imageTranslateX: 0,
      imageTranslateY: 0,
      isDragging: false,
      lastTouchX: 0,
      lastTouchY: 0,
      lastTouchDistance: 0,
      lastTapTime: 0
    };
  },

  watch: {
    visible(val) {
      if (val) {
        this.$nextTick(() => {
          setTimeout(() => this.fitToScreen(), 100);
        });
      } else {
        this.resetTransform();
      }
    }
  },

  methods: {
    handleClose() {
      this.$emit('close');
    },

    onCloseTouch(event) {
      event.stopPropagation();
      this.handleClose();
    },

    handleTap(event) {
      event.stopPropagation();
      const now = Date.now();
      if (this.lastTapTime && (now - this.lastTapTime) < 300) {
        this.imageScale === 1 ? this.zoomToPoint(2, event) : this.resetTransform();
      }
      this.lastTapTime = now;
    },

    handleMpTap(event) {
      event.stopPropagation();
      const now = Date.now();
      if (this.lastTapTime && (now - this.lastTapTime) < 300) {
        this.imageScale === 1 ? this.zoomTo(2) : this.resetTransform();
      }
      this.lastTapTime = now;
    },

    handleLongPress() {
      this.$emit('longpress');
    },

    onZoomIn() {
      this.imageScale = Math.min(this.imageScale * 1.5, 50);
      this.constrainTranslation();
    },

    onZoomOut() {
      this.imageScale = Math.max(this.imageScale / 1.5, 0.5);
      this.constrainTranslation();
    },

    onReset() {
      this.resetTransform();
    },

    resetTransform() {
      this.imageScale = 1;
      this.imageTranslateX = 0;
      this.imageTranslateY = 0;
      this.isDragging = false;
    },

    zoomTo(scale) {
      this.imageScale = scale;
      this.constrainTranslation();
    },

    zoomToPoint(scale, event) {
      const rect = event.currentTarget.getBoundingClientRect
        ? event.currentTarget.getBoundingClientRect()
        : { width: 300, height: 400 };
      const centerX = rect.width / 2;
      const centerY = rect.height / 2;
      const clickX = event.clientX - rect.left;
      const clickY = event.clientY - rect.top;

      this.imageTranslateX -= (clickX - centerX) * (scale - this.imageScale) / this.imageScale;
      this.imageTranslateY -= (clickY - centerY) * (scale - this.imageScale) / this.imageScale;
      this.imageScale = scale;
      this.constrainTranslation();
    },

    constrainTranslation() {
      if (this.imageScale <= 1) {
        this.imageTranslateX = 0;
        this.imageTranslateY = 0;
        return;
      }
      const maxT = 100 * (this.imageScale - 1);
      this.imageTranslateX = Math.max(-maxT, Math.min(maxT, this.imageTranslateX));
      this.imageTranslateY = Math.max(-maxT, Math.min(maxT, this.imageTranslateY));
    },

    onTouchStart(event) {
      event.preventDefault();
      const t = event.touches;
      if (t.length === 1) {
        this.isDragging = true;
        this.lastTouchX = t[0].clientX;
        this.lastTouchY = t[0].clientY;
      } else if (t.length === 2) {
        this.lastTouchDistance = this.getDistance(t[0], t[1]);
      }
    },

    onTouchMove(event) {
      event.preventDefault();
      const t = event.touches;
      if (t.length === 1 && this.isDragging && this.imageScale > 1) {
        this.imageTranslateX += (t[0].clientX - this.lastTouchX) / this.imageScale;
        this.imageTranslateY += (t[0].clientY - this.lastTouchY) / this.imageScale;
        this.constrainTranslation();
        this.lastTouchX = t[0].clientX;
        this.lastTouchY = t[0].clientY;
      } else if (t.length === 2) {
        const dist = this.getDistance(t[0], t[1]);
        this.imageScale = Math.max(0.5, Math.min(50, this.imageScale * (dist / this.lastTouchDistance)));
        this.constrainTranslation();
        this.lastTouchDistance = dist;
      }
    },

    onTouchEnd(event) {
      this.isDragging = false;
    },

    onMouseDown(event) {
      if (this.imageScale > 1) {
        event.preventDefault();
        this.isDragging = true;
        this.lastTouchX = event.clientX;
        this.lastTouchY = event.clientY;
      }
    },

    onMouseMove(event) {
      if (this.isDragging && this.imageScale > 1) {
        event.preventDefault();
        this.imageTranslateX += (event.clientX - this.lastTouchX) / this.imageScale;
        this.imageTranslateY += (event.clientY - this.lastTouchY) / this.imageScale;
        this.constrainTranslation();
        this.lastTouchX = event.clientX;
        this.lastTouchY = event.clientY;
      }
    },

    onMouseUp() {
      this.isDragging = false;
    },

    onWheel(event) {
      const delta = event.deltaY > 0 ? 0.9 : 1.1;
      this.imageScale = Math.max(0.5, Math.min(50, this.imageScale * delta));
      this.constrainTranslation();
    },

    getDistance(t1, t2) {
      const dx = t1.clientX - t2.clientX;
      const dy = t1.clientY - t2.clientY;
      return Math.sqrt(dx * dx + dy * dy);
    },

    fitToScreen() {
      this.resetTransform();

      // #ifdef H5
      const img = new Image();
      img.onload = () => {
        try {
          const container = document.querySelector('.image-preview-container');
          if (!container) return;
          const cw = container.getBoundingClientRect().width - 20;
          const ch = container.getBoundingClientRect().height - 80;
          const isLong = img.naturalHeight > img.naturalWidth * 1.5;
          this.imageScale = isLong
            ? Math.max((cw / img.naturalWidth) * 1.5, 3.0)
            : Math.max(cw / img.naturalWidth, 2.2);
        } catch {
          this.imageScale = 2.2;
        }
      };
      img.src = this.imageUrl;
      // #endif

      // #ifdef MP-WEIXIN
      try {
        uni.getImageInfo({
          src: this.imageUrl,
          success: (res) => {
            const isLong = res.height > res.width * 1.5;
            this.imageScale = isLong ? 3.0 : 2.2;
          },
          fail: () => {
            this.imageScale = 2.2;
          }
        });
      } catch {
        this.imageScale = 2.2;
      }
      // #endif

      // #ifndef MP-WEIXIN && !H5
      this.imageScale = 2.2;
      // #endif
    }
  }
};
</script>

<style scoped>
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.9);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-preview-container {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.image-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  user-select: none;
}

.preview-image {
  max-width: 95%;
  max-height: 95%;
  border-radius: 8px;
  transition: transform 0.2s ease-out;
  pointer-events: auto;
  -webkit-user-select: none;
  user-select: none;
}

.preview-close-btn {
  position: absolute;
  top: 30px;
  right: 20px;
  width: 40px;
  height: 40px;
  background-color: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10001;
  cursor: pointer;
}

.close-icon {
  color: #fff;
  font-size: 28px;
  line-height: 1;
  font-weight: bold;
}

.preview-controls {
  position: absolute;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 10px;
  z-index: 10001;
  background-color: rgba(0, 0, 0, 0.5);
  border-radius: 25px;
  padding: 8px 16px;
}

.control-btn {
  width: 40px;
  height: 40px;
  background-color: rgba(0, 0, 0, 0.5);
  border: none;
  border-radius: 50%;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.control-btn:active {
  background-color: rgba(0, 0, 0, 0.8);
  transform: scale(0.95);
}

.control-icon {
  font-size: 18px;
  font-weight: bold;
}

.control-text {
  font-size: 12px;
}

.zoom-info {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  min-width: 50px;
}

.mp-image-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.mp-preview-image {
  max-width: 90%;
  max-height: 90%;
  border-radius: 8px;
  transition: transform 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes zoomIn {
  from { transform: scale(0.8); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
</style>
