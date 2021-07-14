extern "C"
{
#include <TH/TH.h>
#include <luaT.h>
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
}

#include <glob.h>
#include <iostream>
#include <algorithm>
#include <fstream>
#include <string>
#include <vector>
#include <thread>
#include <unordered_set>

#include "TorchCraft/include/replayer.h"

using namespace torchcraft::replayer;

int main(int argc, char *argv[])
{
    if (argc < 2)
    {
        printf("Usage: %s <replay-files>\n", argv[0]);
        return 1;
    }

    std::vector<std::string> files(argv + 1, argv + argc);
    printf("Processing %d file(s)\n", files.size());

#pragma omp parallel for
    for (int k = 0; k < files.size(); k++)
    {
        auto fname = files[k];

        printf("processing file %s\n", fname);

        std::ifstream inRep(fname);
        Replayer r;
        inRep >> r;

        auto map = r.getRawMap();
        auto x = THByteTensor_size(map, 0);
        auto y = THByteTensor_size(map, 1);
        auto t = r.size();

        std::unordered_set<int32_t> n_units[2];
        int32_t c_ore[2], c_gas[2], n_ore[2], n_gas[2], n_psi[2], n_max_psi[2];
        c_ore[0] = 50;
        c_ore[1] = 50;
        c_gas[0] = 0;
        c_gas[1] = 0;
        n_ore[0] = 0;
        n_ore[1] = 0;
        n_gas[0] = 0;
        n_gas[1] = 0;
        n_psi[0] = 0;
        n_psi[1] = 0;
        n_max_psi[0] = 0;
        n_max_psi[1] = 0;

        for (int i = 0; i < r.size(); i++)
        {
            auto f = r.getFrame(i);
            for (auto team : f->units)
            {
                if (!(team.first == 0 || team.first == 1))
                    continue;
                for (auto unit : team.second)
                {
                    n_units[team.first].insert(unit.id);
                }
            }

            for (int32_t team = 0; team <= 1; team++)
                if (f->resources.find(team) != f->resources.end())
                {
                    auto res = f->resources[team];
                    n_ore[team] += std::max(0, res.ore - c_ore[team]);
                    n_gas[team] += std::max(0, res.gas - c_gas[team]);

                    if (res.used_psi != 0)
                        n_psi[team] = res.used_psi;
                    n_max_psi[team] = std::max(res.used_psi, n_max_psi[team]);

                    c_ore[team] = res.ore;
                    c_gas[team] = res.gas;
                }
        }

        auto walk = THByteTensor_new();
        auto gh = THByteTensor_new();
        auto build = THByteTensor_new();
        std::vector<int> tx, ty;
        r.getMap(walk, gh, build, tx, ty);

#pragma omp critical
        std::cout << "stats for: " << fname;
        std::cout << " " << r.size();
        std::cout << " " << n_units[0].size() << " " << n_units[1].size();
        std::cout << " " << n_ore[0] << " " << n_gas[0] << " " << n_psi[0] << " " << n_max_psi[0];
        std::cout << " " << n_ore[1] << " " << n_gas[1] << " " << n_psi[1] << " " << n_max_psi[1];
        std::cout << " " << x << " " << y << " " << THByteTensor_sumall(walk) << " " << THByteTensor_sumall(gh);
        std::cout << std::endl;
    }
    return 0;
}
