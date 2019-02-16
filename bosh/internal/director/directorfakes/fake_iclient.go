// Code generated by counterfeiter. DO NOT EDIT.
package directorfakes

import (
	"sync"

	"github.com/EngineerBetter/concourse-up/bosh/internal/director"
)

type FakeIClient struct {
	CleanupStub        func() error
	cleanupMutex       sync.RWMutex
	cleanupArgsForCall []struct {
	}
	cleanupReturns struct {
		result1 error
	}
	cleanupReturnsOnCall map[int]struct {
		result1 error
	}
	PathInWorkingDirStub        func(string) string
	pathInWorkingDirMutex       sync.RWMutex
	pathInWorkingDirArgsForCall []struct {
		arg1 string
	}
	pathInWorkingDirReturns struct {
		result1 string
	}
	pathInWorkingDirReturnsOnCall map[int]struct {
		result1 string
	}
	SaveFileToWorkingDirStub        func(string, []byte) (string, error)
	saveFileToWorkingDirMutex       sync.RWMutex
	saveFileToWorkingDirArgsForCall []struct {
		arg1 string
		arg2 []byte
	}
	saveFileToWorkingDirReturns struct {
		result1 string
		result2 error
	}
	saveFileToWorkingDirReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeIClient) Cleanup() error {
	fake.cleanupMutex.Lock()
	ret, specificReturn := fake.cleanupReturnsOnCall[len(fake.cleanupArgsForCall)]
	fake.cleanupArgsForCall = append(fake.cleanupArgsForCall, struct {
	}{})
	fake.recordInvocation("Cleanup", []interface{}{})
	fake.cleanupMutex.Unlock()
	if fake.CleanupStub != nil {
		return fake.CleanupStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.cleanupReturns
	return fakeReturns.result1
}

func (fake *FakeIClient) CleanupCallCount() int {
	fake.cleanupMutex.RLock()
	defer fake.cleanupMutex.RUnlock()
	return len(fake.cleanupArgsForCall)
}

func (fake *FakeIClient) CleanupCalls(stub func() error) {
	fake.cleanupMutex.Lock()
	defer fake.cleanupMutex.Unlock()
	fake.CleanupStub = stub
}

func (fake *FakeIClient) CleanupReturns(result1 error) {
	fake.cleanupMutex.Lock()
	defer fake.cleanupMutex.Unlock()
	fake.CleanupStub = nil
	fake.cleanupReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeIClient) CleanupReturnsOnCall(i int, result1 error) {
	fake.cleanupMutex.Lock()
	defer fake.cleanupMutex.Unlock()
	fake.CleanupStub = nil
	if fake.cleanupReturnsOnCall == nil {
		fake.cleanupReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.cleanupReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeIClient) PathInWorkingDir(arg1 string) string {
	fake.pathInWorkingDirMutex.Lock()
	ret, specificReturn := fake.pathInWorkingDirReturnsOnCall[len(fake.pathInWorkingDirArgsForCall)]
	fake.pathInWorkingDirArgsForCall = append(fake.pathInWorkingDirArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("PathInWorkingDir", []interface{}{arg1})
	fake.pathInWorkingDirMutex.Unlock()
	if fake.PathInWorkingDirStub != nil {
		return fake.PathInWorkingDirStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.pathInWorkingDirReturns
	return fakeReturns.result1
}

func (fake *FakeIClient) PathInWorkingDirCallCount() int {
	fake.pathInWorkingDirMutex.RLock()
	defer fake.pathInWorkingDirMutex.RUnlock()
	return len(fake.pathInWorkingDirArgsForCall)
}

func (fake *FakeIClient) PathInWorkingDirCalls(stub func(string) string) {
	fake.pathInWorkingDirMutex.Lock()
	defer fake.pathInWorkingDirMutex.Unlock()
	fake.PathInWorkingDirStub = stub
}

func (fake *FakeIClient) PathInWorkingDirArgsForCall(i int) string {
	fake.pathInWorkingDirMutex.RLock()
	defer fake.pathInWorkingDirMutex.RUnlock()
	argsForCall := fake.pathInWorkingDirArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeIClient) PathInWorkingDirReturns(result1 string) {
	fake.pathInWorkingDirMutex.Lock()
	defer fake.pathInWorkingDirMutex.Unlock()
	fake.PathInWorkingDirStub = nil
	fake.pathInWorkingDirReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeIClient) PathInWorkingDirReturnsOnCall(i int, result1 string) {
	fake.pathInWorkingDirMutex.Lock()
	defer fake.pathInWorkingDirMutex.Unlock()
	fake.PathInWorkingDirStub = nil
	if fake.pathInWorkingDirReturnsOnCall == nil {
		fake.pathInWorkingDirReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.pathInWorkingDirReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeIClient) SaveFileToWorkingDir(arg1 string, arg2 []byte) (string, error) {
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.saveFileToWorkingDirMutex.Lock()
	ret, specificReturn := fake.saveFileToWorkingDirReturnsOnCall[len(fake.saveFileToWorkingDirArgsForCall)]
	fake.saveFileToWorkingDirArgsForCall = append(fake.saveFileToWorkingDirArgsForCall, struct {
		arg1 string
		arg2 []byte
	}{arg1, arg2Copy})
	fake.recordInvocation("SaveFileToWorkingDir", []interface{}{arg1, arg2Copy})
	fake.saveFileToWorkingDirMutex.Unlock()
	if fake.SaveFileToWorkingDirStub != nil {
		return fake.SaveFileToWorkingDirStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.saveFileToWorkingDirReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeIClient) SaveFileToWorkingDirCallCount() int {
	fake.saveFileToWorkingDirMutex.RLock()
	defer fake.saveFileToWorkingDirMutex.RUnlock()
	return len(fake.saveFileToWorkingDirArgsForCall)
}

func (fake *FakeIClient) SaveFileToWorkingDirCalls(stub func(string, []byte) (string, error)) {
	fake.saveFileToWorkingDirMutex.Lock()
	defer fake.saveFileToWorkingDirMutex.Unlock()
	fake.SaveFileToWorkingDirStub = stub
}

func (fake *FakeIClient) SaveFileToWorkingDirArgsForCall(i int) (string, []byte) {
	fake.saveFileToWorkingDirMutex.RLock()
	defer fake.saveFileToWorkingDirMutex.RUnlock()
	argsForCall := fake.saveFileToWorkingDirArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeIClient) SaveFileToWorkingDirReturns(result1 string, result2 error) {
	fake.saveFileToWorkingDirMutex.Lock()
	defer fake.saveFileToWorkingDirMutex.Unlock()
	fake.SaveFileToWorkingDirStub = nil
	fake.saveFileToWorkingDirReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeIClient) SaveFileToWorkingDirReturnsOnCall(i int, result1 string, result2 error) {
	fake.saveFileToWorkingDirMutex.Lock()
	defer fake.saveFileToWorkingDirMutex.Unlock()
	fake.SaveFileToWorkingDirStub = nil
	if fake.saveFileToWorkingDirReturnsOnCall == nil {
		fake.saveFileToWorkingDirReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.saveFileToWorkingDirReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeIClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.cleanupMutex.RLock()
	defer fake.cleanupMutex.RUnlock()
	fake.pathInWorkingDirMutex.RLock()
	defer fake.pathInWorkingDirMutex.RUnlock()
	fake.saveFileToWorkingDirMutex.RLock()
	defer fake.saveFileToWorkingDirMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeIClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ director.IClient = new(FakeIClient)
