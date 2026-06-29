/**
 * Community Terms Agreement Composable
 * Manages the community terms/agreement state and persistence.
 * Agreement is checked and saved both locally and on the server.
 */

import { request } from '@/utils/request.js'
import { agreeCommunityTerms, getCommunityTermsStatus } from '@/pages/api/community.js'

const TERMS_KEY = 'community_terms_agreed'
const TERMS_VERSION_KEY = 'community_terms_version'
const CURRENT_VERSION = '1.0'

/**
 * Check if the user has agreed to the community terms (local storage only).
 * @returns {boolean}
 */
export function hasAgreedToTermsLocal() {
	try {
		const agreed = uni.getStorageSync(TERMS_KEY)
		const version = uni.getStorageSync(TERMS_VERSION_KEY)
		return agreed === true && version === CURRENT_VERSION
	} catch (e) {
		return false
	}
}

/**
 * Mark the community terms as agreed locally.
 */
export function agreeToTermsLocal() {
	try {
		uni.setStorageSync(TERMS_KEY, true)
		uni.setStorageSync(TERMS_VERSION_KEY, CURRENT_VERSION)
	} catch (e) {
		console.error('Failed to save terms agreement locally:', e)
	}
}

/**
 * Reset terms agreement locally (for testing or re-showing the modal).
 */
export function resetTermsAgreement() {
	try {
		uni.removeStorageSync(TERMS_KEY)
		uni.removeStorageSync(TERMS_VERSION_KEY)
	} catch (e) {
		console.error('Failed to reset terms agreement:', e)
	}
}

/**
 * Save community terms agreement to the server.
 * @returns {Promise<boolean>} true if successful
 */
export async function agreeCommunityTermsToServer() {
	try {
		const res = await agreeCommunityTerms()
		if (!res || !res.success) {
			return false
		}
		return true
	} catch (error) {
		console.error('保存社区须知到服务器失败:', error)
		return false
	}
}

/**
 * Check if the user has agreed to community terms on the server.
 * @returns {Promise<boolean>}
 */
export async function checkCommunityTermsFromServer() {
	try {
		const res = await getCommunityTermsStatus()
		if (!res || !res.success) {
			return false
		}
		return !!(res.result && res.result.agreed)
	} catch (error) {
		// 404 or network error → treat as not agreed
		return false
	}
}

/**
 * Check if user has agreed to terms (local cache first, then server fallback).
 * On first load or if local cache is empty, checks server.
 *
 * @returns {Promise<boolean>} true if already agreed
 */
let _cachedAgreed = null
export async function hasAgreedToTerms() {
	// Local cache hit - fast path
	if (_cachedAgreed !== null) {
		return _cachedAgreed
	}

	// Check local storage first
	const localAgreed = hasAgreedToTermsLocal()
	if (localAgreed) {
		_cachedAgreed = true
		return true
	}

	// Try server check
	try {
		const serverAgreed = await checkCommunityTermsFromServer()
		if (serverAgreed) {
			agreeToTermsLocal()
			_cachedAgreed = true
			return true
		}
	} catch (e) {
		// Server check failed, fall through
	}

	_cachedAgreed = false
	return false
}

/**
 * Agree to community terms - saves both locally and to server.
 * @returns {Promise<boolean>} true if successful
 */
export async function agreeToTerms() {
	try {
		agreeToTermsLocal()
		_cachedAgreed = true
		const serverOk = await agreeCommunityTermsToServer()
		return serverOk
	} catch (error) {
		console.error('同意社区须知失败:', error)
		_cachedAgreed = null
		return false
	}
}

/**
 * Require community terms agreement before proceeding.
 * If terms not agreed, shows the terms modal and returns false.
 *
 * @param {Function} showTermsModal - Callback to show the terms modal
 * @returns {boolean} - true if already agreed, false if modal was shown
 */
export function requireCommunityTerms(showTermsModal) {
	if (hasAgreedToTermsLocal()) {
		return true
	}
	if (typeof showTermsModal === 'function') {
		showTermsModal()
	}
	return false
}

/**
 * Navigate to community only if terms are agreed.
 *
 * @param {Function} showTermsModal - Callback to show the terms modal
 * @returns {boolean} - true if navigation can proceed, false if blocked
 */
export function navigateToCommunityWithTerms(showTermsModal) {
	if (hasAgreedToTermsLocal()) {
		return true
	}
	if (typeof showTermsModal === 'function') {
		showTermsModal()
	}
	return false
}
