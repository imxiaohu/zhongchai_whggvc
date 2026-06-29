jest.mock('./errorHandler.js', () => ({
	handleRequestError: jest.fn(),
	shouldAutoRefreshData: jest.fn(() => false)
}))

jest.mock('./schoolServerStatus.js', () => ({
	handleServerMaintenanceError: jest.fn()
}))

function createUniMock() {
	const store = new Map()
	let networkType = 'wifi'

	return {
		__setNetworkType(t) {
			networkType = t
		},
		getStorageSync: jest.fn((k) => store.get(k)),
		setStorageSync: jest.fn((k, v) => {
			store.set(k, v)
		}),
		removeStorageSync: jest.fn((k) => {
			store.delete(k)
		}),
		getNetworkType: jest.fn(({ success, fail }) => {
			try {
				success({ networkType })
			} catch (e) {
				fail && fail(e)
			}
		}),
		addInterceptor: jest.fn(),
		showToast: jest.fn(),
		reLaunch: jest.fn(),
		request: jest.fn(),
	}
}

describe('utils/request', () => {
	beforeEach(() => {
		global.uni = createUniMock()
		jest.resetModules()
	})

	afterEach(() => {
		delete global.uni
		jest.resetModules()
	})

	it('persists token for /scloud/login response', async () => {
		const { request } = await import('./request.js')
		uni.request.mockImplementation(({ success }) => {
			success({
				statusCode: 200,
				data: { success: true, code: 200, message: 'ok', result: { token: 'T' } },
				header: {}
			})
		})

		const res = await request({ url: '/scloud/login', method: 'POST', data: { username: 'u', password: 'p' } })
		expect(res.result.token).toBe('T')
		expect(uni.setStorageSync).toHaveBeenCalledWith('token', 'T')
	})

	it('returns cached GET response when offline', async () => {
		const { request } = await import('./request.js')
		uni.request.mockImplementation(({ success }) => {
			success({
				statusCode: 200,
				data: { success: true, code: 200, message: 'ok', result: { v: 1 } },
				header: {}
			})
		})

		const first = await request({ url: '/api/m/test', method: 'GET', data: { a: 1 }, cacheTTL: 60 })
		expect(first.result.v).toBe(1)
		expect(uni.request).toHaveBeenCalledTimes(1)

		uni.__setNetworkType('none')
		uni.request.mockClear()

		const second = await request({ url: '/api/m/test', method: 'GET', data: { a: 1 }, cacheTTL: 60 })
		expect(second.fromCache).toBe(true)
		expect(uni.request).toHaveBeenCalledTimes(0)
	})
})
