package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	InvalidMemoryRequestToLimitPercentErr = errors.New("invalid value for MemoryRequestToLimitPercent, must be [0...100]")
	InvalidCPURequestToLimitPercentErr = errors.New("invalid value for CPURequestToLimitPercent, must be [0...100]")
	InvalidLimitCPUToMemoryPercentErr = errors.New("invalid value for LimitCPUToMemoryPercent, must be a positive value")
)

func (c *ClusterResourceOverrideConfig) String() string {
	return fmt.Sprint("MemoryRequestToLimitPercent=%d, CPURequestToLimitPercent=%d LimitCPUToMemoryPercent=%d",
		c.MemoryRequestToLimitPercent, c.CPURequestToLimitPercent, c.LimitCPUToMemoryPercent)
}

func (c *ClusterResourceOverrideConfig) Validate() error {
	if c.MemoryRequestToLimitPercent < 0 || c.MemoryRequestToLimitPercent > 100 {
		return InvalidMemoryRequestToLimitPercentErr
	}

	if c.CPURequestToLimitPercent < 0 || c.CPURequestToLimitPercent > 100 {
		return InvalidCPURequestToLimitPercentErr
	}

	if c.LimitCPUToMemoryPercent < 0 {
		return InvalidLimitCPUToMemoryPercentErr
	}

	return nil
}

func (c *ClusterResourceOverrideConfig) Hash() string {
	value := fmt.Sprintf("%s", c)

	writer := sha256.New()
	writer.Write([]byte(value))
	return hex.EncodeToString(writer.Sum(nil))
}
