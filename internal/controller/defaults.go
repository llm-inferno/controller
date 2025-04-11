package controller

import "time"

// REST server environment variables names
const RestHostEnvName = "INFERNO_HOST"
const RestPortEnvName = "INFERNO_PORT"

// REST server API
var OptimizerURL string

const AddAccelerator = "/addAccelerator"
const RemoveAccelerator = "/removeAccelerator"

const AddModel = "/addModel"
const RemoveModel = "/removeModel"

const AddModelAcceleratorPerf = "/addModelAcceleratorPerf"
const RemoveModelAcceleratorPerf = "/removeModelAcceleratorPerf"

const AddServiceClass = "/addServiceClass"
const RemoveServiceClass = "/removeServiceClass"

const AddServiceClassModelTarget = "/addServiceClassModelTarget"
const RemoveServiceClassModelTarget = "/removeServiceClassModelTarget"

const AddServer = "/addServer"
const RemoveServer = "/removeServer"

const SetCapacities = "/setCapacities"
const RemoveCapacity = "/removeCapacity"

const Optimize = "/optimize"
const ApplyAllocation = "/applyAllocation"

// Others
var RetrialDuration = 30 * time.Second * 30

const FinalizerName = "inferno.neurops.ai/finalizer"
