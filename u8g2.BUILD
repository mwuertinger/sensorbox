load(
    "@platformio_rules//platformio:platformio.bzl",
    "platformio_library",
)

#cc_library(
#    name = "u8g2_cc",
#    srcs = [
#        "cppsrc/U8g2lib.cpp",
#        "cppsrc/U8x8lib.cpp",
#    ],
#    hdrs = [
#        "U8g2lib.h",
#        "U8x8lib.h",
#    ],
#    visibility = ["//visibility:public"],
#)

platformio_library(
    name = "u8g2",
    src = "cppsrc/U8g2lib.cpp",
    hdr = "cppsrc/U8g2lib.h",
    visibility = ["//visibility:public"],
)
