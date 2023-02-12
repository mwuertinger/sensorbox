load(
    "@platformio_rules//platformio:platformio.bzl",
    "platformio_library",
)

platformio_library(
    name = "nanopb",
    hdr = "pb_encode.h",
    deps = ["nanopb_lib"],
    visibility = ["//visibility:public"],
)

cc_library(
    name = "nanopb_lib",
    srcs = [
        "pb_common.c",
	"pb_decode.c",
	"pb_encode.c",
    ],
    hdrs = [
        "pb.h",
	"pb_common.h",
	"pb_decode.h",
	"pb_encode.h",
    ],
)
