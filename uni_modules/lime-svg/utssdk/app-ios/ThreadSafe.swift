import Foundation

@objc public class ThreadSafe: NSObject {

    @objc public static func synchronized(_ lock: NSRecursiveLock, _ block: () -> Void) {
        lock.lock()
        block()
        lock.unlock()
    }
}
