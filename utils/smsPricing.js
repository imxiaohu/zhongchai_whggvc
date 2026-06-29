/**
 * SMS Pricing Utilities
 * SMS package definitions, formatters, and calculators
 */

// SMS Package Definitions
export const PACKAGES = [
	{
		id: 1,
		name: '基础套餐',
		description: '50条短信',
		amount: 500, // 5.00元 (单位：分)
		smsCount: 50,
		pricePerSms: 0.10
	},
	{
		id: 2,
		name: '标准套餐',
		description: '120条短信',
		amount: 1000, // 10.00元
		smsCount: 120,
		pricePerSms: 0.083,
		badge: 'hot'
	},
	{
		id: 3,
		name: '推荐套餐',
		description: '280条短信',
		amount: 2000, // 20.00元
		smsCount: 280,
		pricePerSms: 0.071,
		badge: 'recommended'
	},
	{
		id: 4,
		name: '超值套餐',
		description: '600条短信',
		amount: 4000, // 40.00元
		smsCount: 600,
		pricePerSms: 0.067,
		badge: 'best-value'
	}
];

// Payment Method Definitions
export const PAYMENT_METHODS = {
	wechat: {
		id: 'wechat',
		name: '微信支付',
		icon: 'wechat-pay',
		minAmount: 1
	},
	alipay: {
		id: 'alipay',
		name: '支付宝',
		icon: 'alipay',
		minAmount: 1
	}
};

// Price Formatters
export function formatPrice(cents) {
	if (typeof cents !== 'number') {
		cents = parseInt(cents) || 0;
	}
	return (cents / 100).toFixed(2);
}

export function formatPriceWithUnit(cents) {
	return `¥${formatPrice(cents)}`;
}

export function formatSmsCount(balance, costPerSms = 10) {
	if (typeof balance !== 'number') {
		balance = parseInt(balance) || 0;
	}
	return Math.floor(balance / costPerSms);
}

// Discount Calculators
export function calculateDiscount(originalPrice, discountedPrice) {
	if (originalPrice <= 0) return 0;
	const discount = ((originalPrice - discountedPrice) / originalPrice) * 100;
	return Math.round(discount);
}

export function calculateSavings(package1, package2) {
	if (!package1 || !package2) return 0;
	const price1PerSms = package1.amount / package1.smsCount;
	const price2PerSms = package2.amount / package2.smsCount;
	const savings = (price1PerSms - price2PerSms) * package2.smsCount;
	return Math.max(0, savings);
}

export function getBestValuePackage(packages = PACKAGES) {
	if (!packages || packages.length === 0) return null;
	return packages.reduce((best, current) => {
		const currentPricePerSms = current.amount / current.smsCount;
		const bestPricePerSms = best.amount / best.smsCount;
		return currentPricePerSms < bestPricePerSms ? current : best;
	});
}

export function getRecommendedPackage(packages = PACKAGES) {
	const recommended = packages.find(pkg => pkg.badge === 'recommended');
	return recommended || getBestValuePackage(packages);
}

// Payment Formatters
export function formatPaymentAmount(amount) {
	const amountInt = parseInt(amount);
	if (isNaN(amountInt)) return '0.00';
	return (amountInt / 100).toFixed(2);
}

export function formatPaymentInfo(packageInfo) {
	if (!packageInfo) return '';
	return {
		packageName: packageInfo.name,
		packageDesc: packageInfo.description,
		price: formatPriceWithUnit(packageInfo.amount),
		smsCount: `${packageInfo.smsCount}条`
	};
}

// Order Status Messages
export const ORDER_STATUS = {
	PENDING: 'pending',
	PROCESSING: 'processing',
	SUCCESS: 'success',
	FAILED: 'failed',
	EXPIRED: 'expired',
	CANCELLED: 'cancelled'
};

export function getOrderStatusMessage(status) {
	const messages = {
		[ORDER_STATUS.PENDING]: '等待支付',
		[ORDER_STATUS.PROCESSING]: '支付中...',
		[ORDER_STATUS.SUCCESS]: '支付成功！',
		[ORDER_STATUS.FAILED]: '支付失败，请重试',
		[ORDER_STATUS.EXPIRED]: '订单已过期',
		[ORDER_STATUS.CANCELLED]: '订单已取消'
	};
	return messages[status] || '未知状态';
}

// QR Code Configuration
export const QR_CODE_CONFIG = {
	defaultSize: 200,
	minSize: 100,
	maxSize: 400,
	margin: 1,
	colors: {
		dark: '#000000',
		light: '#FFFFFF'
	}
};

// Payment Timeout Configuration
export const PAYMENT_TIMEOUT = {
	checkInterval: 3000, // 3 seconds
	maxDuration: 300000, // 5 minutes
	successDelay: 2000 // 2 seconds before closing modal
};

// Balance Info Helpers
export function createDefaultBalance() {
	return {
		balance: 0,
		balanceYuan: '0.00',
		smsCost: 10
	};
}

export function updateBalanceFromResponse(balanceData) {
	return {
		balance: balanceData.balance || 0,
		balanceYuan: balanceData.balanceYuan || '0.00',
		smsCost: balanceData.smsCost || 10
	};
}
