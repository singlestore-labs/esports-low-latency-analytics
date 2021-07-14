set (CMAKE_CXX_STANDARD 11)
cmake_minimum_required (VERSION 2.8.9)
project (replay-processor)

find_package(OpenMP)
if (OPENMP_FOUND)
    set (CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OpenMP_C_FLAGS}")
    set (CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${OpenMP_CXX_FLAGS}")
    set (CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} ${OpenMP_EXE_LINKER_FLAGS}")
endif()

link_directories($ENV{HOME}/torch/install/lib)
include_directories($ENV{HOME}/torch/install/include)

include_directories(/StarData)
include_directories(/StarData/TorchCraft/include)

file(GLOB torchcraft_srcs "/StarData/TorchCraft/replayer/*.cpp")

add_executable(replay-processor main.cpp ${torchcraft_srcs})
target_link_libraries(replay-processor $ENV{HOME}/torch/install/lib/libTH.so)