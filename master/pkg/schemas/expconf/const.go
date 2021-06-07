package expconf

// Configuration constants for task name generator.
const (
	TaskNameGeneratorWords = 3
	TaskNameGeneratorSep   = "-"
)

// Default task environment docker image names.
const (
	DefaultCPUImage = "determinedai/environments:py-3.7-pytorch-1.8-lightning-1.2-tf-2.4-cpu-125e7dc"
	DefaultGPUImage = "determinedai/environments:cuda-11.1-pytorch-1.8-lightning-1.2-tf-2.4-gpu-0.9.0"
)
